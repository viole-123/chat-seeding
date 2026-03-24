# Uniscore Seeding Bot

## Overview

Uniscore Seeding Bot là một hệ thống tự động tạo và gửi tin nhắn giả lập (bot) vào các phòng chat trận đấu nhằm tăng tương tác và tạo bầu không khí sôi động cho trận cầu. Bot có thể phản ứng theo các sự kiện real-time như bàn thắng, thẻ phạt, thay người,... và tạo ra các bình luận mang tính cá nhân hóa dựa trên "nhân vật" (persona) được chọn.

## Features

- Phản ứng real-time với các sự kiện trận đấu (goal, card, substitution,...)
- Hỗ trợ nhiều kiểu nhân vật (casual, funny, analyst,...) với ngôn ngữ/tone khác nhau
- Bộ lọc chất lượng nội dung (anti-spam, anti-duplicate, sentiment check)
- Quản lý tần suất gửi tin theo nhiều cấp độ: event-type limit, match limit, bot ratio
- Tích hợp với Kafka (nguồn event), Redis (context/state), PostgreSQL (log), LLM (fallback generation)
- Hỗ trợ xuất bản qua WebSocket hoặc MQTT broker

## Architecture

```mermaid
graph TD
    A[Kafka Producer] --> B[(Kafka)]
    B --> C[Consumer Worker]
    C --> D[Policy Checker]
    D --> E[Persona Selector]
    E --> F[Message Generator]
    F --> G[Quality Filter]
    G --> H[Publish Flow]
    H --> I[WebSocket Client]
    H --> J[MQTT Broker]

    style C fill:#FFE4B5,stroke:#333
    style D fill:#98FB98,stroke:#333
    style E fill:#87CEEB,stroke:#333
    style F fill:#DDA0DD,stroke:#333
    style G fill:#FFA07A,stroke:#333
    style H fill:#FFD700,stroke:#333
```

### Core Components

#### 1. Consumer Layer
- **Kafka Consumer**: Nhận event trận đấu từ Kafka topic.
- **Mapper**: Chuyển event Kafka sang model nội bộ.

#### 2. Processing Layer
- **Policy Checker**: Áp dụng các rule rate limit, bot ratio, kill-switch,...
- **Persona Selector**: Chọn nhân vật phù hợp với ngữ cảnh trận.
- **Context Builder**: Tổng hợp thông tin trận, lịch sử chat, intent người dùng.
- **Message Generator**: Sinh nội dung từ template hoặc LLM.
- **Quality Filter**: Đánh giá và lọc nội dung không đạt tiêu chuẩn.

#### 3. Output Layer
- **Publisher**: Gửi tin ra WebSocket hoặc MQTT broker.
- **Shadow Ban**: Ngăn không xuất bản ra ngoài khi cần thiết.

#### 4. Storage
- **Redis**: Lưu trạng thái trận, phòng chat, cooldown, persona state,...
- **PostgreSQL**: Log tin nhắn đã gửi để audit/debug.

#### 5. External Services
- **LLM Gateway**: Kết nối với LLM (local hoặc cloud) để fallback khi template không đủ.
- **Metrics**: Prometheus endpoint để giám sát.

## Technologies

- Go 1.21+
- Kafka (event streaming)
- Redis (in-memory store)
- PostgreSQL (persistent logs)
- Docker & Docker Compose (dev environment)
- Gin (HTTP server for metrics)
- Sarama (Kafka client)
- Redis/go-redis (Redis client)
- Viper (config management)

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Kafka cluster (local or remote)
- Redis instance
- PostgreSQL database

### Configuration

File cấu hình mẫu: `config.yaml`

```yaml
service_name: uniscore-seeding-bot

kafka:
  brokers:
    - localhost:9092
  topic: seeding-events
  username: ""
  password: ""

redis_matches:
  addr: localhost:6379
  username: "default"
  password: "" 
  db: 0

database:
  url: "postgres://postgres:postgres@localhost:5432/uniscore_seeding?sslmode=disable"

vllm:
  api_url: "http://localhost:11434"
  model: "gpt-oss:120b-cloud"
  timeout: 90s

seeding_policy:
  max_messages_bot: 100
  bot_ratio: 0.3
  cooldown: 60s
  max_events_per_hour: 60
  dedup_window: 24h
  enable_kill_switch: true

quality:
  min_length: 10
  max_length: 200
  banned_words: ["fuck", "shit", "damn"]
  dedup_ttl: 300
```

### Run Locally

```bash
# 1. Clone repo
git clone https://github.com/your-org/uniscore-seeding-bot.git
cd uniscore-seeding-bot

# 2. Install dependencies
go mod tidy

# 3. Start dependencies (Kafka, Redis, PostgreSQL)
docker-compose -f deployments/docker/docker-compose.yml up -d

# 4. Run bot
go run ./cmd/seeding-bot

# Or build binary
go build -o bin/seeding-bot ./cmd/seeding-bot
./bin/seeding-bot
```

### Deployment

Docker image sẵn sàng trong `deployments/docker/Dockerfile`.

```bash
# Build image
docker build -t uniscore/seeding-bot .

# Run container
docker run -d \
  --name seeding-bot \
  -v $(pwd)/config.yaml:/app/config.yaml \
  uniscore/seeding-bot
```


## Monitoring

Metrics Prometheus được export tại `http://localhost:8080/metrics`.

Các metric chính:
- `seeding_events_total`
- `seeding_messages_published_total`
- `seeding_quality_filter_rejected_total`
- `seeding_shadowban_skipped_total`

## Contributing

1. Fork repo
2. Tạo branch feature: `git checkout -b feature/NewFeature`
3. Commit thay đổi: `git commit -m 'Add new feature'`
4. Push lên branch: `git push origin feature/NewFeature`
5. Tạo Pull Request

## License

MIT
