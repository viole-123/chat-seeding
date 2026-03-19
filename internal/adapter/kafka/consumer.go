package kafka

import (
	"context"
	"log"
	"time"
	"uniscore-seeding-bot/internal/config"
	"uniscore-seeding-bot/internal/domain/service"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer     sarama.ConsumerGroup
	contextStore service.ContextStore
	topic        string
}

type MConsumerGroup struct {
	config       *config.KafkaConfig
	topic        string
	group        sarama.ConsumerGroup
	logger       *log.Logger
	contextStore service.ContextStore
	lastErrLogAt time.Time
}

func NewConsumer(cfg *config.KafkaConfig, groupId string, topic string, contextStore service.ContextStore) (*MConsumerGroup, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	if cfg.Username != "" && cfg.Password != "" {
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.User = cfg.Username
		saramaConfig.Net.SASL.Password = cfg.Password
	}
	group, err := sarama.NewConsumerGroup(cfg.Brokers, groupId, saramaConfig)
	if err != nil {
		return nil, err
	}
	logger := log.Default()
	logger.Println("✅ success create consumer group")
	go func() {
		for err := range group.Errors() {
			logger.Println("consumererror:", err)
		}
	}()
	return &MConsumerGroup{config: cfg, topic: topic, group: group, logger: logger, contextStore: contextStore}, nil
}

func (c *MConsumerGroup) RegisterHandlerAndConsumeMessage(ctx context.Context, handler sarama.ConsumerGroupHandler) {
	defer c.group.Close()
	for {
		select {
		case <-ctx.Done():
			c.logger.Println("context done, stopping consumer")
			return
		default:
		}
		if err := c.group.Consume(ctx, []string{c.topic}, handler); err != nil {
			if time.Since(c.lastErrLogAt) > 30*time.Second {
				c.logger.Printf("error from consumer (throttled): %v", err)
				c.lastErrLogAt = time.Now()
			}
			time.Sleep(2 * time.Second)
		}
	}
}
