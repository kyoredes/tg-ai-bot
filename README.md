# RageAI

Бот с AI на базе микросервисной архитектуры. Бекенд написан на Go и Python.

## Возможности

- Общение с AI-моделью в свободном формате
- Напоминания: пользователь создаёт событие через чат, бот сохраняет его и отправляет уведомление в нужное время
- Система подписок и лимитов запросов

## Архитектура

Gateway оркестрирует все вызовы — сервисы не знают друг о друге и не вызывают друг друга напрямую.

```
aiogram (Python)
      │  ▲
      │  │ напоминания
      ▼  │
   gateway (Go)
      ├──▶ auth-service         # 1. проверка токена
      ├──▶ subscription-service # 2. проверка лимитов
      ├──▶ ai-service           # 3. запрос к AI-модели
      └──▶ reminder-service     # создание / управление событиями
```

Сервисы общаются через **gRPC**. Auth-токены кешируются в Redis.

### Флоу напоминаний

```
Пользователь → "Напомни завтра в 10:00 позвонить маме"
      ↓
ai-service извлекает дату, время и текст из сообщения
      ↓
reminder-service сохраняет событие в PostgreSQL
      ↓
планировщик срабатывает в 10:00
      ↓
aiogram отправляет сообщение пользователю
```

## Сервисы

| Сервис | Язык | Описание |
|---|---|---|
| `services/bot` | Python, aiogram 3 | Telegram-интерфейс, приём и отправка сообщений |
| `services/gateway` | Go, chi | Оркестрация вызовов, JWT middleware |
| `services/auth-service` | Go | Регистрация, авторизация, выдача JWT |
| `services/subscription-service` | Go | Тарифные планы, лимиты запросов |
| `services/ai-service` | Python, FastAPI | Интеграция с AI-моделью, streaming-ответы |
| `services/reminder-service` | Go | Хранение событий, планировщик уведомлений |

## Структура репозитория

```
rageai/
├── services/
│   ├── bot/
│   ├── gateway/
│   ├── auth-service/
│   ├── subscription-service/
│   ├── ai-service/
│   └── reminder-service/
├── proto/               # gRPC контракты (.proto файлы)
├── deploy/
│   ├── docker-compose.yml
│   └── docker-compose.prod.yml
├── scripts/
│   └── gen-proto.sh     # кодогенерация из .proto
├── .env.example
└── README.md
```

## Инфраструктура

- **PostgreSQL** — отдельная база на каждый сервис
- **Redis** — общий, кеширование токенов и rate limiting
- **Docker Compose** — локальная разработка
- **nginx / Traefik** — TLS termination перед gateway в продакшне

## Локальный запуск

Требования: Docker, Docker Compose, Go 1.22+, Python 3.12+

```bash
git clone https://github.com/you/rageai.git
cd rageai

cp .env.example .env
# заполни .env своими ключами

docker compose -f deploy/docker-compose.yml up --build
```

## Переменные окружения

```env
TELEGRAM_TOKEN=

AI_API_KEY=
AI_MODEL=

JWT_SECRET=

POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_DB=rageai

REDIS_URL=redis://redis:6379
```

Файл `.env` добавлен в `.gitignore`.

## Разработка

```bash
# Кодогенерация gRPC из .proto файлов
./scripts/gen-proto.sh

# Запустить один сервис локально
cd services/gateway && go run ./cmd/main.go

# Тесты
cd services/auth-service && go test ./...
```

## CI/CD

GitHub Actions запускает деплой только изменённых сервисов через path filters:

```yaml
on:
  push:
    paths:
      - 'services/gateway/**'
      - 'proto/**'
```

## Порядок реализации

1. `auth-service` — регистрация, JWT
2. `gateway` — оркестрация с заглушками
3. `bot` — базовый диалог
4. `subscription-service` — планы и лимиты
5. `ai-service` — интеграция с AI, streaming
6. `reminder-service` — события и планировщик

## Лицензия

MIT
