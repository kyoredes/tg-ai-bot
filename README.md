> 🇷🇺 [Русская версия / Russian version](README.ru.md)

# AI Bot

An AI-powered Telegram bot built on a microservices architecture. Backend services are written in Go and Python.

## Features

- Free-form AI chat via Telegram
- User profile and subscription info
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
      └──▶ ai-service            # LLM requests
```

Services communicate over **gRPC**. Auth tokens are cached in Redis.

### AI chat flow

```
User sends a message in Telegram
      ↓
aiogram-service → gateway-service
      ↓
auth-service (profile) + subscription-service (limits)
      ↓
ai-service (OpenAI in prod, G4F as fallback / G4F only in dev)
      ↓
Response delivered back to the user
```

## Services

| Service | Stack | Description |
|---|---|---|
| `aiogram-service` | Python, aiogram 3 | Telegram bot UI |
| `gateway-service` | Go, Gin | HTTP API, gRPC orchestration |
| `auth-service` | Go | Registration, auth, JWT |
| `subscription-service` | Go | Plans and request limits |
| `ai-service` | Python, gRPC | LLM integration (OpenAI + G4F) |

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
- **Redis** — token cache and AI chat history
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

# ai-service: DEBUG=true → G4F only, false → OpenAI + G4F fallback
AI_DEBUG=true
OPENAI_API_KEY=
OPENAI_BASE_URL=https://api.openai.com/v1
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
