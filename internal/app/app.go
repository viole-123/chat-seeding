package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"uniscore-seeding-bot/internal/adapter/kafka"
	"uniscore-seeding-bot/internal/adapter/mqtt"
	"uniscore-seeding-bot/internal/adapter/postgres"
	"uniscore-seeding-bot/internal/adapter/publisher"
	"uniscore-seeding-bot/internal/adapter/redis"
	"uniscore-seeding-bot/internal/adapter/vllm"
	"uniscore-seeding-bot/internal/config"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/observability/metrics"
	"uniscore-seeding-bot/internal/pkg/circuitbreaker"
	"uniscore-seeding-bot/internal/transport/http/handler"
	"uniscore-seeding-bot/internal/usecase/safety"
	"uniscore-seeding-bot/internal/usecase/seeding"
	"uniscore-seeding-bot/internal/usecase/template"
	"uniscore-seeding-bot/internal/worker"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

// Bootstrap initializes services and starts background workers.
func Bootstrap() error {
	ctx := context.Background()

	//1 load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	log.Printf("✅ config loaded: %+v\n", cfg)
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if err := metrics.Init(); err != nil {
		return fmt.Errorf("init metrics failed: %w", err)
	}
	//2 init services
	redisClient, err := redis.NewRedisClient(ctx, cfg.RedisMatches)
	if err != nil {
		return fmt.Errorf("connect redis failed :%w", err)
	}
	defer redisClient.Close()
	log.Println("✅ redis all connected")

	//3 start background workers
	//go startMatchWorker(ctx, redisClient)
	//go startLeagueWorker(ctx, redisClient)
	depupService := redis.NewDedupService(redisClient, 24*time.Hour)
	//go startDedupWorker(ctx, depupService)
	killSwitchService := seeding.NewKillSwitchService(redisClient)
	shadowBanService := safety.NewShadowBanService(redisClient)
	//go startKillSwitchWorker(ctx, killSwitchService)
	rateLimitService := redis.NewRateLimit(redisClient)
	//go startRateLimitWorker(ctx, rateLimitService)
	contextStore := redis.NewContextStoreService(redisClient)
	roomManager := seeding.NewRoomManager(redisClient)

	log.Println("✅ Services initialized")
	qualityStateService := redis.NewQualityStateService(redisClient)
	qualityFilter := seeding.NewQualityFilter(qualityStateService, cfg.Quality)
	log.Println("✅ Quality filter initialized")
	personaStateService := redis.NewPersonaStateService(redisClient)
	personaSelector, err := seeding.NewPersonaSelector(personaStateService, "personas.yaml")
	if err != nil {
		return fmt.Errorf("init persona selector failed: %w", err)
	}
	log.Println("✅ Persona selector initialized")
	policyChecker := seeding.NewPolicyChecker(
		depupService,
		killSwitchService,
		rateLimitService,
		cfg.SeedingPolicy,
		personaStateService,
		contextStore,
	)
	log.Println("✅ Policy checker initialized")
	consumer, err := kafka.NewConsumer(&cfg.KafkaConfig, cfg.KafkaConfig.Topic+"_group", cfg.KafkaConfig.Topic, contextStore)
	if err != nil {
		return fmt.Errorf("failed to create Kafka consumer: %w", err)
	}
	log.Println("✅ Kafka consumer created")
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("connect postgres failed: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Println("✅ Postgres connected")
	templateRepo := postgres.NewTemplateRepo(db)
	templateLoader := template.NewTemplateLoader(templateRepo)
	messageRepo := postgres.NewMessageRepo(db)
	log.Println("✅ Message repository initialized")
	log.Println("✅ Template loader initialized")
	templateRender := template.NewTemplateRenderer()
	log.Println("✅ Template renderer initialized")
	llmGateway := vllm.NewVLLMGateway(cfg.VLLM.APIURL, cfg.VLLM.Model, cfg.VLLM.Timeout)
	log.Println("✅ LLM Gateway initialized")
	intentDetector := seeding.NewIntentDetector(llmGateway)
	log.Println("✅ Intent detector initialized")
	contextBuilder := seeding.NewContextBuilder(contextStore, intentDetector)
	log.Println("✅ Context builder initialized")
	messageGenerator := seeding.NewMessageGenerator(templateLoader, templateRender, llmGateway)
	log.Println("✅ Message generator initialized")
	botReplySystem := seeding.NewBotReplySystem(intentDetector, llmGateway, personaSelector)
	log.Println("✅ Bot reply system initialized")
	//auto
	autoScalerImpl, err := redis.NewAutoScalerImpl(redisClient)
	if err != nil {
		return fmt.Errorf("failed to create auto scaler impl: %w", err)
	}
	autoScaler := seeding.NewAutoScalerLogic(autoScalerImpl)
	log.Println("✅ AutoScaling inintialized)")

	scalerWorker := worker.NewScalerWorker(contextStore, autoScaler, 60*time.Second, log.Default())
	go scalerWorker.Start(ctx)
	log.Println("✅ Scaler worker started (interval: 60s)")
	gatewayURL := "ws://localhost:8080/ws"
	basePublisher := publisher.NewWebSocketPublisher(gatewayURL)
	basePublisher.SetBroadcaster(handler.BroadcastMessage)
	circuitBreaker := circuitbreaker.New()
	publisherService := publisher.NewCircuitBreakerPublisher(basePublisher, circuitBreaker)
	log.Println("✅ Publisher initialized (WebSocket + CircuitBreaker)")

	mqttBrokerURL := strings.TrimSpace(os.Getenv("MQTT_BROKER_URL"))
	if mqttBrokerURL == "" {
		mqttBrokerURL = "tcp://localhost:1883"
	}

	var mqttPublisher *mqtt.Publisher
	mqttPublisher, err = mqtt.NewPublisher(mqttBrokerURL, fmt.Sprintf("seeding-bot-pub-%d", time.Now().UnixNano()))
	if err != nil {
		log.Printf("⚠️  [MQTT] publisher init failed, fallback to websocket publish: %v", err)
		mqttPublisher = nil
	} else {
		log.Printf("✅ [MQTT] publisher connected: %s", mqttBrokerURL)
	}

	eventHandler := seeding.NewEventHandler(*policyChecker, personaSelector, contextBuilder, contextStore, messageGenerator, qualityFilter, log.Default(), publisherService, mqttPublisher, roomManager, depupService, messageRepo, autoScaler)
	eventHandler.SetShadowBanService(shadowBanService)
	log.Println("✅ Event handler initialized")
	prematchHandler := seeding.NewPrematchHandler(
		contextStore,
		contextBuilder,
		messageGenerator,
		personaSelector,
		publisherService,
		mqttPublisher,
		roomManager,
		messageRepo,
		log.Default(),
		autoScaler,
	)
	log.Println("✅ Prematch handler initialized")

	handler.SetUserMessageHandler(func(in handler.IncomingUserMessage) {
		matchID := in.MatchID
		if matchID == "" {
			return
		}
		roomID := in.RoomID
		if roomID == "" || roomID == matchID {
			resolvedRoomID, roomErr := roomManager.GetOrCreate(context.Background(), matchID)
			if roomErr != nil {
				roomID = fmt.Sprintf("room-%s", matchID)
			} else {
				roomID = resolvedRoomID
			}
		}

		userMessage := model.ChatMessage{
			ID:        fmt.Sprintf("user-%d", time.Now().UnixNano()),
			Content:   in.Content,
			Timestamp: time.Now().Unix(),
			IsBot:     false,
			MatchID:   matchID,
			RoomID:    roomID,
			EventType: "USER_MESSAGE",
			CreatedAt: time.Now(),
		}

		if mqttPublisher != nil {
			mqttMsg := mqtt.ChatMessage{
				ID:        userMessage.ID,
				MatchID:   matchID,
				RoomID:    roomID,
				UserID:    "dashboard-user",
				Content:   userMessage.Content,
				Timestamp: userMessage.Timestamp,
				IsBot:     false,
				EventType: userMessage.EventType,
			}
			if err := mqttPublisher.PublishUserMessage(context.Background(), roomID, mqttMsg); err != nil {
				log.Printf("⚠️  [USER-MSG] mqtt publish failed: %v", err)
			}
		} else {
			if err := publisherService.Publish(userMessage); err != nil {
				log.Printf("⚠️  [USER-MSG] websocket publish failed: %v", err)
			}
		}
		_ = contextStore.PushChatMessage(roomID, userMessage)
		if messageRepo != nil {
			_ = messageRepo.SaveMessage(userMessage)
		}

		bundle, err := contextBuilder.BuildBundle(context.Background(), matchID, roomID)
		if err != nil || bundle == nil {
			bundle = &model.ContextBundle{Match: model.MatchState{MatchID: matchID}}
		}
		bundle.Chat.RawMessages = append([]model.ChatMessage{userMessage}, bundle.Chat.RawMessages...)

		replyCtx, cancelReply := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancelReply()
		reply, err := botReplySystem.GenerateReply(replyCtx, model.UserMessage{Content: in.Content}, *bundle)
		if err != nil || reply == nil || reply.ReplyType == model.ReplyTypeSkip || reply.Text == "" {
			return
		}

		botMsg := model.ChatMessage{
			ID:        fmt.Sprintf("reply-%d", time.Now().UnixNano()),
			Content:   reply.Text,
			Timestamp: time.Now().Unix(),
			IsBot:     true,
			Persona:   reply.PersonaID,
			MatchID:   matchID,
			RoomID:    roomID,
			EventType: "USER_REPLY",
			CreatedAt: time.Now(),
		}

		if mqttPublisher != nil {
			mqttMsg := mqtt.ChatMessage{
				ID:        botMsg.ID,
				MatchID:   matchID,
				RoomID:    roomID,
				UserID:    reply.PersonaID,
				Content:   botMsg.Content,
				Timestamp: botMsg.Timestamp,
				IsBot:     true,
				PersonaID: reply.PersonaID,
				EventType: botMsg.EventType,
			}
			if err := mqttPublisher.PublishBotMessage(context.Background(), roomID, mqttMsg); err != nil {
				log.Printf("⚠️  [BOT-REPLY] mqtt publish failed: %v", err)
				return
			}
		} else {
			if err := publisherService.Publish(botMsg); err != nil {
				log.Printf("⚠️  [BOT-REPLY] publish failed: %v", err)
				return
			}
		}

		_ = contextStore.PushChatMessage(roomID, botMsg)
		if messageRepo != nil {
			_ = messageRepo.SaveMessage(botMsg)
		}
	})
	consumerCtx, cancelConsumer := context.WithCancel(ctx)

	if mqttPublisher != nil {
		bridge, bridgeErr := mqtt.NewConsumerBridge(
			mqttBrokerURL,
			fmt.Sprintf("seeding-bot-bridge-%d", time.Now().UnixNano()),
			handler.BroadcastMessageToRoom,
		)
		if bridgeErr != nil {
			log.Printf("⚠️  [MQTT] consumer bridge init failed: %v", bridgeErr)
		} else {
			go func() {
				if err := bridge.Start(consumerCtx); err != nil {
					log.Printf("⚠️  [MQTT] bridge stopped: %v", err)
				}
			}()
			log.Println("✅ [MQTT] consumer bridge started (room/+ -> websocket room) ")
		}
	}

	go func() {
		log.Println("Starting kafka consumer...")
		consumer.RegisterHandlerAndConsumeMessage(consumerCtx, eventHandler)
	}()
	adminHandler := handler.NewAdminHandler(killSwitchService, shadowBanService)
	activeMatchesHandler := handler.NewActiveMatchesHandler(contextStore)
	messageHistoryHandler := handler.NewMessageHistoryHandler(messageRepo)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, ngrok-skip-browser-warning")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	router.Static("/web", "./web")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/")
	})

	router.GET("/ws", handler.WebSocketHandler)
	router.GET("/matches/active", activeMatchesHandler.List)
	router.GET("/messages/history", messageHistoryHandler.List)
	router.GET("/metrics", gin.WrapF(metrics.Handler))
	// Test api/chat/send
	router.POST("/api/chat/send", func(c *gin.Context) {
		var in handler.IncomingUserMessage
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if in.MatchID == "" || in.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "match_id and content are required"})
			return
		}
		msg := mqtt.ChatMessage{
			ID:        fmt.Sprintf("user-%d", time.Now().UnixNano()),
			RoomID:    in.RoomID,
			Content:   in.Content,
			Timestamp: time.Now().Unix(),
			IsBot:     false, // user thật
		}
		mqttPublisher.PublishBotMessage(ctx, in.RoomID, msg)
		c.JSON(http.StatusAccepted, gin.H{
			"ok":       true,
			"message":  "message is being processed",
			"match_id": in.MatchID,
			"room_id":  in.RoomID,
		})
	})

	log.Println("✅ WebSocket endpoint registered at /ws")
	producerCfg := sarama.NewConfig()
	producerCfg.Producer.Return.Successes = true
	producerCfg.Producer.RequiredAcks = sarama.WaitForAll
	kafkaProducer, err := sarama.NewSyncProducer(cfg.KafkaConfig.Brokers, producerCfg)
	if err != nil {
		log.Printf("⚠️  [EVENT-SENDER] Kafka producer init failed (endpoint disabled): %v", err)
	} else {
		defer kafkaProducer.Close()
		eventSenderHandler := handler.NewEventSenderHandler(kafkaProducer, cfg.KafkaConfig.Topic)
		router.POST("/send-event", eventSenderHandler.SendEvent)
		log.Println("✅ Event sender endpoint registered at POST /send-event")
	}
	prematchPoller := worker.NewPrematchPoller(
		prematchHandler,
		5*time.Minute,
		log.Default(),
		contextStore,
	)

	go prematchPoller.Start(consumerCtx)
	// log.Println("✅ Prematch poller started (interval: 5m)")
	admin := router.Group("/admin")
	{
		admin.GET("/kill-switch/status", adminHandler.GetKillSwitch)
		admin.POST("/kill-switch", adminHandler.SetKillSwitch)
		admin.GET("/shadow-ban/status", adminHandler.GetShadowBan)
		admin.POST("/shadow-ban", adminHandler.SetShadowBan)
	}
	httpPort := strings.TrimSpace(os.Getenv("HTTP_PORT"))
	if httpPort == "" {
		httpPort = "8081"
	}
	httpAddr := ":" + httpPort
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}
	go func() {
		log.Printf("🚀 Starting HTTP server on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// 10. Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down gracefully...")

	// 11. Graceful shutdown
	cancelConsumer()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println("✅ Shutdown complete")
	return nil

}
