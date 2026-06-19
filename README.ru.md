# AI Bot

Бот с AI на базе микросервисной архитектуры. Бекенд написан на Go и Python.

## Возможности

- Общение с AI-моделью в свободном формате через Telegram
- Профиль и информация о подписке
- Система подписок и лимитов запросов (в разработке)
- Напоминания: пользователь создаёт событие через чат, бот сохраняет его и отправляет уведомление в нужное время (планируется)

## Архитектура

Gateway оркестрирует все вызовы — сервисы не знают друг о друге и не вызывают друг друга напрямую.

```
aiogram-service (Python)
      │
      ▼
gateway-service (Go)
      ├──▶ auth-service         # регистрация, профиль
      ├──▶ subscription-service # тарифы и лимиты
      └──▶ ai-service           # запрос к LLM
```

Сервисы общаются через **gRPC**. Auth-токены кешируются в Redis.

### Флоу чата с AI

```
Пользователь → сообщение в Telegram
      ↓
aiogram-service → gateway-service
      ↓
auth-service (профиль) + subscription-service (лимиты)
      ↓
ai-service (OpenAI в prod, G4F как fallback / только G4F в dev)
      ↓
Ответ пользователю в Telegram
```

## Сервисы

| Сервис | Язык | Описание |
|---|---|---|
| `aiogram-service` | Python, aiogram 3 | Telegram-интерфейс |
| `gateway-service` | Go, Gin | HTTP API, оркестрация gRPC-вызовов |
| `auth-service` | Go | Регистрация, авторизация, JWT |
| `subscription-service` | Go | Тарифные планы, лимиты запросов |
| `ai-service` | Python, gRPC | Интеграция с LLM (OpenAI + G4F) |

## Структура репозитория

```
ai-bot/
├── aiogram-service/
├── gateway-service/
├── auth-service/
├── subscription-service/
├── ai-service/
├── proto/               # gRPC контракты (.proto)
├── deploy/
│   └── docker-compose.yml
├── scripts/
│   └── gen-proto.sh     # кодогенерация из .proto
├── .env.example
├── README.md
└── README.ru.md
```

## Инфраструктура

- **PostgreSQL** — отдельная база на auth и subscription
- **Redis** — кеширование токенов и история чата AI
- **Docker Compose** — локальная разработка

## Локальный запуск

Требования: Docker, Docker Compose

```bash
git clone <repository-url>
cd ai-bot

cp deploy/.env.example deploy/.env
# заполни BOT_TOKEN и при необходимости OPENAI_API_KEY

docker compose -f deploy/docker-compose.yml up --build
```

## Переменные окружения

Основной файл для Docker: `deploy/.env`

```env
BOT_TOKEN=
COMMON_PUB_KEY=secret
JWT_SECRET_KEY=dev-jwt-secret-key-change-in-prod!!

# ai-service: DEBUG=true — только G4F, false — OpenAI + G4F fallback
AI_DEBUG=true
OPENAI_API_KEY=
OPENAI_BASE_URL=https://api.openai.com/v1
```

Каждый сервис также может иметь свой `.env` для локального запуска вне Docker.

## Разработка

```bash
# Кодогенерация gRPC из .proto
./scripts/gen-proto.sh

# Тесты auth-service
cd auth-service && go test ./...
```

## Лицензия

MIT
