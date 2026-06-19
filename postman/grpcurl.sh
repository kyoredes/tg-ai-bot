#!/usr/bin/env bash
# Примеры gRPC-запросов через grpcurl (альтернатива Postman)
# Установка: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

TELEGRAM_ID="${TELEGRAM_ID:-123456789}"
USER_ID="${USER_ID:-}"

echo "=== auth-service: StartTelegram ==="
grpcurl -plaintext \
  -d "{\"telegram_id\": \"${TELEGRAM_ID}\"}" \
  localhost:50051 \
  auth.v1.AuthService/StartTelegram

echo ""
echo "=== subscription-service: GetSubscriptionByUserId ==="
if [[ -z "${USER_ID}" ]]; then
  echo "Задай USER_ID из ответа auth-service, например:"
  echo "USER_ID=... ./postman/grpcurl.sh"
  exit 0
fi

grpcurl -plaintext \
  -d "{\"user_id\": \"${USER_ID}\"}" \
  localhost:50052 \
  subscription.v1.SubscriptionService/GetSubscriptionByUserId
