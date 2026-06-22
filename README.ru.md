# AI Bot

Бот с AI на базе микросервисной архитектуры. Бекенд написан на Go и Python.

## Возможности

- Общение с AI-моделью в свободном формате через Telegram
- Разбор Telegram-профиля (асинхронно через Kafka)
- Профиль и информация о подписке
- Админ-панель: история чатов, промпты, health
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
      ├──▶ ai-service           # запрос к LLM (gRPC + Kafka worker)
      └──▶ Redpanda (Kafka)     # асинхронный разбор профиля
```

Сервисы общаются через **gRPC**. Auth-токены кешируются в Redis. Разбор профиля — **асинхронный** через Kafka (Redpanda).

### Флоу чата с AI (синхронный)

```
Пользователь → сообщение в Telegram
      ↓
aiogram-service → gateway-service
      ↓
auth-service (профиль) + subscription-service (лимиты)
      ↓
ai-service (LiteLLM в prod, G4F как fallback / только G4F в dev)
      ↓
Ответ пользователю в Telegram
```

### Флоу разбора профиля (асинхронный)

```
Пользователь нажимает «Разбор профиля»
      ↓
aiogram-service → gateway-service  (HTTP 202 Accepted, мгновенно)
      ↓
Kafka: profile.analyze.requests
      ↓
ai-service worker (LLM в фоне)
      ↓
Kafka: profile.analyze.results
      ↓
aiogram-service consumer → ответ в чат
```

Бот сразу показывает «Разбор начат» и присылает результат, когда worker закончит. Долгих HTTP-таймаутов нет.

## Сервисы

| Сервис | Язык | Описание |
|---|---|---|
| `aiogram-service` | Python, aiogram 3 | Telegram-интерфейс, Kafka consumer результатов |
| `gateway-service` | Go, Gin | HTTP API, оркестрация gRPC, Kafka producer |
| `auth-service` | Go | Регистрация, авторизация, JWT |
| `subscription-service` | Go | Тарифные планы, лимиты запросов |
| `ai-service` | Python, gRPC, Kafka | LLM (LiteLLM + G4F), worker разбора профиля |
| `admin-panel` | React | Админка: сессии, промпты, health |

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
- **Redis** — кеширование токенов, история чата и разборов профиля
- **Redpanda** — Kafka-совместимый брокер для асинхронного разбора профиля
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

# ai-service: DEBUG=true — только G4F, false — LiteLLM + G4F fallback
AI_DEBUG=true
LITELLM_API_KEY=
LITELLM_API_BASE=
OPENAI_API_KEY=
OPENAI_BASE_URL=https://api.openai.com/v1

# Kafka (Redpanda в Docker Compose)
KAFKA_BROKERS=redpanda:9092

# HTTP-таймауты (только чат; разбор профиля асинхронный)
GATEWAY_TIMEOUT=120s
HTTP_TIMEOUT=120
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
