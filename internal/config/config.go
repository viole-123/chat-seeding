package config

import (
	"fmt"
	"log"
	"os"
	"time"
	"uniscore-seeding-bot/internal/domain/model"

	"github.com/goccy/go-yaml"
)

// Config holds application configuration.
type Config struct {
	ServiceName   string                   `yaml:"service_name"`
	KafkaConfig   KafkaConfig              `yaml:"kafka"`
	Redis         RedisConfig              `yaml:"redis"`
	Database      DatabaseConfig           `yaml:"database"`
	VLLM          VLLMConfig               `yaml:"vllm"`
	SeedingPolicy SeedingPolicy            `yaml:"seeding_policy"`
	Quality       model.QualityCheckConfig `yaml:"quality"`
	RedisMatches  RedisConfig              `yaml:"redis_matches"`
}

// ⭐ THÊM VLLMConfig
type VLLMConfig struct {
	APIURL  string        `yaml:"api_url"`
	Model   string        `yaml:"model"`
	Timeout time.Duration `yaml:"timeout"`
}
type DatabaseConfig struct {
	URL string `yaml:"url"`
}

// Load loads configuration (env/file).
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}
	//read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}
	//log ok
	log.Printf("✅ Config loaded from %s", path)
	log.Printf("   ServiceName: %s", cfg.ServiceName)
	log.Printf("   Kafka Topic: %s", cfg.KafkaConfig.Topic)
	if cfg.RedisMatches.Addr != "" {
		log.Printf("   Redis Matches: %s", cfg.RedisMatches.Addr)
	} else {
		log.Printf("   Redis: %s", cfg.Redis.Addr)
	}
	log.Printf("   Database: %s", cfg.Database.URL)
	log.Printf("   VLLM API URL: %s", cfg.VLLM.APIURL)
	log.Printf("   VLLM Model: %s", cfg.VLLM.Model)
	log.Printf("   VLLM Timeout: %s", cfg.VLLM.Timeout)
	log.Printf("   MaxMessagesBot: %d", cfg.SeedingPolicy.MaxMessagesBot)
	log.Printf("   BotRatio: %.2f", cfg.SeedingPolicy.BotRatio)
	log.Printf("   Cooldown: %s", cfg.SeedingPolicy.Cooldown)
	log.Printf("   Quality MinLength: %d", cfg.Quality.MinLength)
	log.Printf("   Quality MaxLength: %d", cfg.Quality.MaxLength)
	log.Printf("   Banned Words: %d items", len(cfg.Quality.BannedWords))
	return &cfg, nil
}
