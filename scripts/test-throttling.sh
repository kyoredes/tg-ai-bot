#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/deploy"

echo "Restarting services for clean in-memory rate-limit state..."
docker compose restart gateway-service auth-service subscription-service ai-service >/dev/null
sleep 5

pass=0
fail=0

check() {
  local name="$1"
  local expected="$2"
  local actual="$3"
  if [[ "$actual" == "$expected" ]]; then
    echo "  PASS: $name (got $actual)"
    pass=$((pass + 1))
  else
    echo "  FAIL: $name (expected $expected, got $actual)"
    fail=$((fail + 1))
  fi
}

echo "=== 1. Gateway: admin login throttle (limit 10/min) ==="
login_429=0
login_200=0
for i in $(seq 1 12); do
  code=$(curl -s -o /dev/null -w '%{http_code}' -X POST http://localhost:8000/admin/login \
    -H 'Content-Type: application/json' \
    -d '{"username":"admin","password":"wrong"}')
  if [[ "$code" == "429" ]]; then
    login_429=$((login_429 + 1))
  elif [[ "$code" == "401" || "$code" == "200" ]]; then
    login_200=$((login_200 + 1))
  fi
done
echo "  results: ${login_200}x 401/200, ${login_429}x 429"
if [[ "$login_429" -ge 1 && "$login_200" -ge 1 ]] || [[ "$login_429" -ge 10 ]]; then
  echo "  PASS: login throttled (429 after limit reached)"
  pass=$((pass + 1))
else
  echo "  FAIL: expected 429 after 10 login attempts"
  fail=$((fail + 1))
fi

echo
echo "=== 2. Gateway: telegram chat throttle (limit 10/min per telegramID) ==="
chat_429=0
chat_other=0
for i in $(seq 1 15); do
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 2 -X POST http://localhost:8000/telegram/chat \
    -H 'Content-Type: application/json' \
    -H 'Authorization: secret' \
    -d '{"telegramID":"throttle-only-gateway","prompt":"x"}' 2>/dev/null || echo "000")
  if [[ "$code" == "429" ]]; then
    chat_429=$((chat_429 + 1))
  else
    chat_other=$((chat_other + 1))
  fi
done
echo "  results: ${chat_other}x non-429, ${chat_429}x 429"
if [[ "$chat_429" -ge 3 ]]; then
  echo "  PASS: chat throttled (429 after ~10 attempts)"
  pass=$((pass + 1))
else
  echo "  FAIL: expected 429 after ~10 chat attempts"
  fail=$((fail + 1))
fi

echo
echo "=== 3. auth-service gRPC throttle (limit 300/min) ==="
docker exec agrobot-ai-service sh -c 'cd /app && uv run python << "PY"
import sys
import grpc

channel = grpc.insecure_channel("auth-service:50051")
stub = channel.unary_unary("/auth.v1.AuthService/Health", lambda x: b"", lambda x: x)

throttled = False
ok = 0
for i in range(305):
    try:
        stub(b"")
        ok += 1
    except grpc.RpcError as exc:
        if exc.code() == grpc.StatusCode.RESOURCE_EXHAUSTED:
            throttled = True
            print(f"  throttled at request {i + 1}", flush=True)
            break
        raise

print(f"  results: {ok}x ok, throttled={throttled}", flush=True)
sys.exit(0 if throttled and ok >= 300 else 1)
PY' 2>&1 || auth_exit=$?
if [[ "${auth_exit:-0}" -eq 0 ]]; then
  echo "  PASS: auth gRPC throttled after 300 requests"
  pass=$((pass + 1))
else
  echo "  FAIL: auth gRPC throttle did not trigger"
  fail=$((fail + 1))
fi

echo
echo "=== 4. subscription-service gRPC throttle (limit 300/min) ==="
docker exec agrobot-ai-service sh -c 'cd /app && uv run python << "PY"
import sys
import grpc

channel = grpc.insecure_channel("subscription-service:50052")
stub = channel.unary_unary("/subscription.v1.SubscriptionService/Health", lambda x: b"", lambda x: x)

throttled = False
ok = 0
for i in range(305):
    try:
        stub(b"")
        ok += 1
    except grpc.RpcError as exc:
        if exc.code() == grpc.StatusCode.RESOURCE_EXHAUSTED:
            throttled = True
            print(f"  throttled at request {i + 1}", flush=True)
            break
        raise

print(f"  results: {ok}x ok, throttled={throttled}", flush=True)
sys.exit(0 if throttled and ok >= 300 else 1)
PY' 2>&1 || sub_exit=$?
if [[ "${sub_exit:-0}" -eq 0 ]]; then
  echo "  PASS: subscription gRPC throttled after 300 requests"
  pass=$((pass + 1))
else
  echo "  FAIL: subscription gRPC throttle did not trigger"
  fail=$((fail + 1))
fi

echo
echo "=== 5. ai-service gRPC global throttle (limit 300/min) ==="
docker exec agrobot-ai-service sh -c 'cd /app && uv run python << "PY"
import sys
import grpc

channel = grpc.insecure_channel("127.0.0.1:50053")
stub = channel.unary_unary("/ai.v1.AIService/Health", lambda x: b"", lambda x: x)

throttled = False
ok = 0
for i in range(305):
    try:
        stub(b"")
        ok += 1
    except grpc.RpcError as exc:
        if exc.code() == grpc.StatusCode.RESOURCE_EXHAUSTED:
            throttled = True
            print(f"  throttled at request {i + 1}", flush=True)
            break
        raise

print(f"  results: {ok}x ok, throttled={throttled}", flush=True)
sys.exit(0 if throttled and ok >= 300 else 1)
PY' 2>&1 || ai_exit=$?
if [[ "${ai_exit:-0}" -eq 0 ]]; then
  echo "  PASS: ai-service global gRPC throttled after 300 requests"
  pass=$((pass + 1))
else
  echo "  FAIL: ai-service global gRPC throttle did not trigger"
  fail=$((fail + 1))
fi

echo
echo "=== 6. ai-service Chat throttle (limit 10/min per telegramID) ==="
docker exec agrobot-ai-service sh -c 'cd /app && uv run python << "PY"
import sys
import grpc
from rpc.ai.v1 import ai_pb2, ai_pb2_grpc

channel = grpc.insecure_channel("127.0.0.1:50053")
stub = ai_pb2_grpc.AIServiceStub(channel)

throttled = 0
ok = 0
for i in range(15):
    try:
        stub.Chat(
            ai_pb2.ChatRequest(telegram_id="grpc-throttle-test", prompt="x"),
            timeout=2,
        )
        ok += 1
    except grpc.RpcError as exc:
        if exc.code() == grpc.StatusCode.RESOURCE_EXHAUSTED:
            throttled += 1
        else:
            ok += 1

print(f"  results: {ok}x ok/timeout, {throttled}x throttled", flush=True)
sys.exit(0 if throttled >= 3 else 1)
PY' 2>&1 || ai_chat_exit=$?
if [[ "${ai_chat_exit:-0}" -eq 0 ]]; then
  echo "  PASS: ai-service Chat throttled after ~10 requests"
  pass=$((pass + 1))
else
  echo "  FAIL: ai-service Chat throttle did not trigger"
  fail=$((fail + 1))
fi

echo
echo "=== 7. aiogram-service middleware (limit 5 per 5s) ==="
docker exec -i agrobot-aiogram-service python3 << 'PY'
import asyncio
import sys
from types import SimpleNamespace

from aiogram.dispatcher.event.bases import SkipHandler
from middleware.throttling import ThrottlingMiddleware


class FakeMessage:
    def __init__(self, user_id: int):
        self.from_user = SimpleNamespace(id=user_id)
        self.answers = []

    async def answer(self, text: str):
        self.answers.append(text)


async def main():
    mw = ThrottlingMiddleware(limit_count=5, limit_seconds=5.0, ban_seconds=100.0)
    msg = FakeMessage(999001)
    skipped = 0
    passed = 0

    async def handler(event, data):
        return "ok"

    for i in range(7):
        try:
            result = await mw(handler, msg, {})
            if result == "ok":
                passed += 1
        except SkipHandler:
            skipped += 1

    print(f"  results: {passed}x passed, {skipped}x skipped", flush=True)
    ok = passed == 5 and skipped == 2
    sys.exit(0 if ok else 1)


asyncio.run(main())
PY
2>&1 || aio_exit=$?
if [[ "${aio_exit:-0}" -eq 0 ]]; then
  echo "  PASS: aiogram middleware blocks after 5 events"
  pass=$((pass + 1))
else
  echo "  FAIL: aiogram middleware did not block as expected"
  fail=$((fail + 1))
fi

echo
echo "=============================="
echo "TOTAL: $pass passed, $fail failed"
if [[ "$fail" -gt 0 ]]; then
  exit 1
fi
