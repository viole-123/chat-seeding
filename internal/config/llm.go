package config

import "time"

type LLMConfig struct {
	APIURL  string        `yaml:"api_url"`
	Model   string        `yaml:"model"`
	Timeout time.Duration ``
}
