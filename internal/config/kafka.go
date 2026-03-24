package config

// KafkaConfig holds Kafka-related configuration.

type KafkaConfig struct {
	Brokers               []string `yaml:"brokers"`
	Username              string   `yaml:"username"`
	Password              string   `yaml:"password"`
	Topic                 string   `yaml:"topic"`
	Retries               int      `yaml:"retries"`
	ProducerReturnSuccess bool     `yaml:"producer_return_success"`
}

type LogConfig struct {
	RotationSize  int `yaml:"rotation_size"`
	RotationCount int `yaml:"rotation_count"`
}

type KafkaConfigStruct struct {
	Kafka         KafkaConfig   `yaml:"kafka"`
	Log           LogConfig     `yaml:"log"`
	SeedingPolicy SeedingPolicy `yaml:"seeding_policy"`
}

// func LoadConfig(configpath string) (*KafkaConfigStruct, error) {
// 	_, err := os.Stat(configpath)
// 	if os.IsNotExist(err) {
// 		log.Fatalf("config file does not exit :%v", err)
// 	}
// 	file, err := os.Open(configpath)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer file.Close()
// 	var cfg KafkaConfigStruct
// 	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
// 		return nil, fmt.Errorf("decode yaml: %w", err)
// 	}
// 	return &cfg, nil
// }
