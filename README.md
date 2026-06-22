> 🇷🇺 [Русская версия / Russian version](README.ru.md)

# AI Bot

An AI-powered Telegram bot built on a microservices architecture. Backend services are written in Go and Python.

## Features

- Free-form AI chat via Telegram
- Telegram profile roast (async via Kafka)
- User profile and subscription info
- Admin panel for chat history and LLM config
- Subscription plans and request limits (in progress)
- Reminders: create events in chat and get notified on schedule (planned)

## Architecture

The gateway orchestrates all calls — services are isolated and never call each other directly.

```
aiogram-service (Python)
      │
      ▼
gateway-service (Go)
      ├──▶ auth-service          # registration, profile
      ├──▶ subscription-service  # plans and limits
      ├──▶ ai-service            # LLM requests (gRPC + Kafka worker)
      └──▶ Redpanda (Kafka)      # async profile analysis jobs
```

Services communicate over **gRPC**. Auth tokens are cached in Redis. Profile analysis is **asynchronous** via Kafka (Redpanda).

### AI chat flow (synchronous)

```
User sends a message in Telegram
      ↓
aiogram-service → gateway-service
      ↓
auth-service (profile) + subscription-service (limits)
      ↓
ai-service (LiteLLM in prod, G4F as fallback / G4F only in dev)
      ↓
Response delivered back to the user
```

### Profile roast flow (asynchronous)

```
User taps "Profile roast" in Telegram
      ↓
aiogram-service → gateway-service  (HTTP 202 Accepted, ~instant)
      ↓
Kafka topic: profile.analyze.requests
      ↓
ai-service worker (LLM analysis in background)
      ↓
Kafka topic: profile.analyze.results
      ↓
aiogram-service consumer → message delivered to the user
```

The bot shows a progress message immediately and sends the roast when the worker finishes. No long HTTP timeouts.

## Services

| Service | Stack | Description |
|---|---|---|
| `aiogram-service` | Python, aiogram 3 | Telegram bot UI, Kafka result consumer |
| `gateway-service` | Go, Gin | HTTP API, gRPC orchestration, Kafka producer |
| `auth-service` | Go | Registration, auth, JWT |
| `subscription-service` | Go | Plans and request limits |
| `ai-service` | Python, gRPC, Kafka | LLM integration (LiteLLM + G4F), profile worker |
| `admin-panel` | React | Admin UI for sessions, prompts, health |

## Repository layout

```
ai-bot/
├── aiogram-service/
├── gateway-service/
├── auth-service/
├── subscription-service/
├── ai-service/
├── proto/               # gRPC contracts (.proto)
├── deploy/
│   └── docker-compose.yml
├── scripts/
│   └── gen-proto.sh     # proto code generation
├── .env.example
├── README.md
└── README.ru.md
```

## Infrastructure

- **PostgreSQL** — separate databases for auth and subscription
- **Redis** — token cache, AI chat history, profile roast history
- **Redpanda** — Kafka-compatible broker for async profile analysis
- **Docker Compose** — local development

## Local setup

Requirements: Docker, Docker Compose

```bash
git clone <repository-url>
cd ai-bot

cp deploy/.env.example deploy/.env
# set BOT_TOKEN and OPENAI_API_KEY if needed

docker compose -f deploy/docker-compose.yml up --build
```

## Environment variables

Main file for Docker: `deploy/.env`

```env
BOT_TOKEN=
COMMON_PUB_KEY=secret
JWT_SECRET_KEY=dev-jwt-secret-key-change-in-prod!!

# ai-service: DEBUG=true → G4F only, false → LiteLLM + G4F fallback
AI_DEBUG=true
LITELLM_API_KEY=
LITELLM_API_BASE=
OPENAI_API_KEY=
OPENAI_BASE_URL=https://api.openai.com/v1

# Kafka (Redpanda in Docker Compose)
KAFKA_BROKERS=redpanda:9092

# HTTP timeouts (chat only; profile roast is async)
GATEWAY_TIMEOUT=120s
HTTP_TIMEOUT=120
```

Each service may also have its own `.env` for running outside Docker.

## Development

```bash
# Generate gRPC code from .proto files
./scripts/gen-proto.sh

# Run auth-service tests
cd auth-service && go test ./...
```

## License

MIT
