package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	"uniscore-seeding-bot/internal/domain/model"

	"github.com/redis/go-redis/v9"
)

type ContextStoreServiceImpl struct {
	redisClient *RedisClient
}

// Init ContextStoreServiceImpl with RedisClient
func NewContextStoreService(redisClient *RedisClient) *ContextStoreServiceImpl {
	return &ContextStoreServiceImpl{
		redisClient: redisClient,
	}
}
func (s *ContextStoreServiceImpl) GetMatchState(matchID string) (model.MatchState, error) {
	ctx := context.Background()
	key := fmt.Sprintf("match:state:%s", matchID)

	log.Printf("🔍 Fetching match state from Redis with key: %s", key)

	data, err := s.redisClient.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return model.MatchState{}, fmt.Errorf("redis HGETALL failed for key %s: %w", key, err)
	}

	// HGetAll không có redis.Nil, nếu key không tồn tại thì map sẽ rỗng
	if len(data) == 0 {
		log.Printf("⚠️  No match state found in Redis for key: %s", key)
		return model.MatchState{
			MatchID: matchID,
		}, nil
	}

	matchState := model.MatchState{
		MatchID: matchID, // ưu tiên matchID truyền vào, đừng phụ thuộc Redis có field này hay không
		RoomID:  getString(data, "room_id", ""),

		HomeTeam: model.Team{
			ID:        getString(data, "home_team_id", ""),
			Name:      getString(data, "home_team_name", ""),
			ShortName: getString(data, "home_team_short", ""),
		},

		AwayTeam: model.Team{
			ID:        getString(data, "away_team_id", ""),
			Name:      getString(data, "away_team_name", ""),
			ShortName: getString(data, "away_team_short", ""),
		},

		Competition: model.Competition{
			ID:   getString(data, "competition_id", ""),
			Name: getString(data, "competition_name", ""),
			Tier: getInt(data, "competition_tier", 0),

			Country: model.Country{
				ID:   getString(data, "country_id", ""),
				Name: getString(data, "country_name", ""),
			},

			Category: model.Category{
				ID:   getString(data, "category_id", ""),
				Name: getString(data, "category_name", ""),
			},
		},

		MatchTime: getInt64(data, "match_time", 0),
		Date:      getString(data, "date", ""),
		Phase:     model.MatchPhase(getString(data, "phase", "prematch")),
		Minute:    getInt(data, "minute", 0),
		HomeScore: getInt(data, "home_score", 0),
		AwayScore: getInt(data, "away_score", 0),
		UpdatedAt: getInt64(data, "updated_at", 0),
	}

	return matchState, nil
}

func getString(data map[string]string, key string, defaultValue string) string {
	if v, ok := data[key]; ok && v != "" {
		return v
	}
	return defaultValue
}

func getInt(data map[string]string, key string, defaultValue int) int {
	v, ok := data[key]
	if !ok || v == "" {
		return defaultValue
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}

func getInt64(data map[string]string, key string, defaultValue int64) int64 {
	v, ok := data[key]
	if !ok || v == "" {
		return defaultValue
	}

	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultValue
	}
	return n
}

func (s *ContextStoreServiceImpl) SetMatchState(matchID string, matchCtx model.MatchState) error {
	ctx := context.Background()
	key := fmt.Sprintf("match:state:%s", matchID)

	fields := map[string]interface{}{
		"match_id":         matchCtx.MatchID,
		"room_id":          matchCtx.RoomID,
		"minute":           fmt.Sprintf("%d", matchCtx.Minute),
		"home_score":       fmt.Sprintf("%d", matchCtx.HomeScore),
		"away_score":       fmt.Sprintf("%d", matchCtx.AwayScore),
		"phase":            string(matchCtx.Phase),
		"match_time":       fmt.Sprintf("%d", matchCtx.MatchTime),
		"date":             matchCtx.Date,
		"updated_at":       fmt.Sprintf("%d", matchCtx.UpdatedAt),
		"home_team_id":     matchCtx.HomeTeam.ID,
		"home_team_name":   matchCtx.HomeTeam.Name,
		"home_team_short":  matchCtx.HomeTeam.ShortName,
		"away_team_id":     matchCtx.AwayTeam.ID,
		"away_team_name":   matchCtx.AwayTeam.Name,
		"away_team_short":  matchCtx.AwayTeam.ShortName,
		"competition_id":   matchCtx.Competition.ID,
		"competition_name": matchCtx.Competition.Name,
		"competition_tier": fmt.Sprintf("%d", matchCtx.Competition.Tier),
		"country_id":       matchCtx.Competition.Country.ID,
		"country_name":     matchCtx.Competition.Country.Name,
		"category_id":      matchCtx.Competition.Category.ID,
		"category_name":    matchCtx.Competition.Category.Name,
	}

	if err := s.redisClient.redis.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("redis hset failed: %w", err)
	}
	if err := s.redisClient.redis.Expire(ctx, key, 6*time.Hour).Err(); err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	return nil
}

func (s *ContextStoreServiceImpl) GetRecentEvents(matchId string, limit int) ([]model.CompactEvent, error) {
	ctx := context.Background()
	key := fmt.Sprintf("match:events:%s", matchId)

	results, err := s.redisClient.redis.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		if err == redis.Nil {
			return []model.CompactEvent{}, nil
		}
		return nil, fmt.Errorf("redis lrange failed :%w", err)
	}
	events := make([]model.CompactEvent, 0, len(results))
	for _, jsonStr := range results {
		var event model.CompactEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

func (s *ContextStoreServiceImpl) PushEvent(matchID string, event model.MatchEvent) error {
	ctx := context.Background()
	key := fmt.Sprintf("match:events:%s", matchID)

	eventJson, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}
	if err := s.redisClient.redis.LPush(ctx, key, eventJson).Err(); err != nil {
		return fmt.Errorf("redis lpush failed: %w", err)
	}

	if err := s.redisClient.redis.LTrim(ctx, key, 0, 49).Err(); err != nil {
		return fmt.Errorf("redis ltrim failed: %w", err)
	}
	if err := s.redisClient.redis.Expire(ctx, key, 6*time.Hour).Err(); err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	return nil
}

func (s *ContextStoreServiceImpl) GetRecentChatWindow(roomID string, limit int) (model.ChatContext, error) {
	ctx := context.Background()
	primaryKey := fmt.Sprintf("match:chat:%s", roomID)
	legacyKey := fmt.Sprintf("conv:window:%s", roomID)

	results, err := s.redisClient.redis.LRange(ctx, primaryKey, 0, int64(limit-1)).Result()
	if err != nil && err != redis.Nil {
		return model.ChatContext{}, fmt.Errorf("redis lrange failed :%w", err)
	}
	if len(results) == 0 {
		legacyResults, legacyErr := s.redisClient.redis.LRange(ctx, legacyKey, 0, int64(limit-1)).Result()
		if legacyErr != nil && legacyErr != redis.Nil {
			return model.ChatContext{}, fmt.Errorf("redis lrange failed :%w", legacyErr)
		}
		results = legacyResults
	}

	messages := decodeChatMessages(results)
	return model.ChatContext{
		RoomID:            roomID,
		LastBotMessages:   toMessageIDs(messages),
		LastMessageHashes: toMessageIDs(messages),
		RawMessages:       messages,
	}, nil
}

func (s *ContextStoreServiceImpl) GetContentChatWindow(roomID string, limit int) (model.ChatContext, error) {
	ctx := context.Background()
	key := fmt.Sprintf("match:chat:%s", roomID)

	results, err := s.redisClient.redis.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		if err == redis.Nil {
			return model.ChatContext{
				RoomID: roomID,
			}, nil
		}
		return model.ChatContext{}, fmt.Errorf("redis lrange failed :%w", err)
	}

	messages := decodeChatMessages(results)
	chatCtx := model.ChatContext{
		RoomID:            roomID,
		LastBotMessages:   toMessageIDs(messages),
		LastMessageHashes: toMessageIDs(messages),
		RawMessages:       messages,
	}
	return chatCtx, nil
} //?

func (s *ContextStoreServiceImpl) PushChatMessage(roomID string, msg model.ChatMessage) error {
	ctx := context.Background()
	key := fmt.Sprintf("match:chat:%s", roomID)
	legacyKey := fmt.Sprintf("conv:window:%s", roomID)

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	if err := s.redisClient.redis.LPush(ctx, key, msgJSON).Err(); err != nil {
		return fmt.Errorf("redis lpush failed: %w", err)
	}
	if err := s.redisClient.redis.LPush(ctx, legacyKey, msgJSON).Err(); err != nil {
		return fmt.Errorf("redis lpush failed: %w", err)
	}
	// LTRIM giữ tối đa 50 messages
	if err := s.redisClient.redis.LTrim(ctx, key, 0, 49).Err(); err != nil {
		return fmt.Errorf("redis ltrim failed: %w", err)
	}
	if err := s.redisClient.redis.LTrim(ctx, legacyKey, 0, 49).Err(); err != nil {
		return fmt.Errorf("redis ltrim failed: %w", err)
	}

	// Set TTL 24h
	if err := s.redisClient.redis.Expire(ctx, key, 6*time.Hour).Err(); err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	if err := s.redisClient.redis.Expire(ctx, legacyKey, 6*time.Hour).Err(); err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}

	return nil
}

func decodeChatMessages(values []string) []model.ChatMessage {
	messages := make([]model.ChatMessage, 0, len(values))
	for _, jsonStr := range values {
		var msg model.ChatMessage
		if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages
}

func toMessageIDs(messages []model.ChatMessage) []string {
	ids := make([]string, 0, len(messages))
	for _, msg := range messages {
		if msg.ID != "" {
			ids = append(ids, msg.ID)
		}
	}
	return ids
}

const matchDailyHashKey = "matches_daily"

func (s *ContextStoreServiceImpl) GetMatchByID(ctx context.Context, matchID string) (*model.MatchDailyCatchFromRedis, error) {
	val, err := s.redisClient.redis.HGet(ctx, matchDailyHashKey, matchID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var match model.MatchDailyCatchFromRedis
	if err := json.Unmarshal([]byte(val), &match); err != nil {
		return nil, err
	}
	return &match, nil
}

func (s *ContextStoreServiceImpl) GetAllTodayMatches(ctx context.Context) ([]*model.MatchDailyCatchFromRedis, error) {
	fields, err := s.redisClient.redis.HGetAll(ctx, matchDailyHashKey).Result()
	if err != nil {
		return nil, err
	}

	log.Printf("🔍 [ContextStore] matches_daily has %d raw fields", len(fields))

	matches := make([]*model.MatchDailyCatchFromRedis, 0, len(fields))
	for field, val := range fields {
		var match model.MatchDailyCatchFromRedis
		if err := json.Unmarshal([]byte(val), &match); err != nil {
			log.Printf("⚠️  [ContextStore] skip field=%s: invalid JSON (%v)", field, err)
			continue
		}
		matches = append(matches, &match)
	}

	return matches, nil
}

func keys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func (s *ContextStoreServiceImpl) HasSentPrematch(ctx context.Context, matchID string) (bool, error) {
	key := fmt.Sprintf("bot:prematch:sent:%s", matchID)
	exists, err := s.redisClient.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *ContextStoreServiceImpl) MarkSentPrematch(ctx context.Context, matchID string) error {
	key := fmt.Sprintf("bot:prematch:sent:%s", matchID)
	return s.redisClient.redis.Set(ctx, key, "1", 2*time.Hour).Err()
}

func (s *ContextStoreServiceImpl) GetBotCount(ctx context.Context, roomID string) (int64, error) {
	key := fmt.Sprintf("room:online:%s", roomID)
	count, err := s.redisClient.redis.SCard(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("redis SCARD %s: %w", key, err)
	}
	return count, nil
}
