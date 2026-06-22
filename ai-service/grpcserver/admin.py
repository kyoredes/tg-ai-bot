import json
import logging

from redis.asyncio import Redis

from config.history import chat_history_settings
from config.llm import litellm_settings
from config.settings import settings
from llm.g4f_models import G4F_FALLBACK_MODELS, G4F_MODELS
from llm.history.store import ChatHistoryStore
from llm.prompt.store import SystemPromptStore

logger = logging.getLogger(__name__)


class AdminService:
    def __init__(self, redis_client: Redis | None = None):
        self._redis = redis_client or Redis.from_url(
            settings.REDIS_URL,
            decode_responses=True,
            socket_timeout=5,
            socket_connect_timeout=5,
        )
        self._prompt_store = SystemPromptStore(self._redis)

    async def get_chat_history(self, telegram_id: str) -> list[dict[str, str]]:
        store = ChatHistoryStore(session_id=telegram_id, redis_client=self._redis)
        await store.ping()
        messages = await store.load()
        return [{"role": m.role, "content": m.content} for m in messages]

    async def clear_chat_history(self, telegram_id: str) -> None:
        store = ChatHistoryStore(session_id=telegram_id, redis_client=self._redis)
        await store.ping()
        await store.clear()

    async def list_chat_sessions(self, page: int, limit: int) -> tuple[list[dict], int]:
        if page < 1:
            page = 1
        if limit < 1:
            limit = 20

        prefix = chat_history_settings.KEY_PREFIX
        keys: list[str] = []
        cursor = 0
        while True:
            cursor, batch = await self._redis.scan(cursor=cursor, match=f"{prefix}*", count=100)
            keys.extend(batch)
            if cursor == 0:
                break

        total = len(keys)
        offset = (page - 1) * limit
        page_keys = keys[offset : offset + limit]

        sessions = []
        for key in page_keys:
            telegram_id = key.removeprefix(prefix)
            raw = await self._redis.get(key)
            message_count = 0
            if raw:
                try:
                    payload = json.loads(raw)
                    if isinstance(payload, list):
                        message_count = len(payload)
                except json.JSONDecodeError:
                    pass
            sessions.append({"telegram_id": telegram_id, "message_count": message_count})

        return sessions, total

    def get_llm_config(self) -> dict:
        uses_litellm = not settings.DEBUG
        g4f_models = tuple(
            str(model) for model in (G4F_MODELS if settings.DEBUG else G4F_FALLBACK_MODELS)
        )
        provider = "g4f" if settings.DEBUG else "litellm+g4f"
        return {
            "model": litellm_settings.MODEL if uses_litellm else "",
            "temperature": litellm_settings.TEMPERATURE if uses_litellm else 0.0,
            "max_tokens": litellm_settings.MAX_TOKENS if uses_litellm else 0,
            "debug": settings.DEBUG,
            "provider": provider,
            "g4f_models": g4f_models,
            "uses_litellm": uses_litellm,
        }

    async def count_chat_sessions(self) -> int:
        prefix = chat_history_settings.KEY_PREFIX
        count = 0
        cursor = 0
        while True:
            cursor, batch = await self._redis.scan(cursor=cursor, match=f"{prefix}*", count=100)
            count += len(batch)
            if cursor == 0:
                break
        return count

    async def get_profile_roast_history(self, telegram_id: str) -> list[dict]:
        from llm.history.profile_roast_store import ProfileRoastStore

        store = ProfileRoastStore(telegram_id=telegram_id, redis_client=self._redis)
        await store.ping()
        entries = await store.load()
        return [entry.model_dump() for entry in entries]

    async def clear_profile_roast_history(self, telegram_id: str) -> None:
        from llm.history.profile_roast_store import ProfileRoastStore

        store = ProfileRoastStore(telegram_id=telegram_id, redis_client=self._redis)
        await store.ping()
        await store.clear()

    async def list_profile_roast_sessions(self, page: int, limit: int) -> tuple[list[dict], int]:
        from config.profile_roast import profile_roast_settings

        if page < 1:
            page = 1
        if limit < 1:
            limit = 20

        prefix = profile_roast_settings.KEY_PREFIX
        keys: list[str] = []
        cursor = 0
        while True:
            cursor, batch = await self._redis.scan(
                cursor=cursor,
                match=f"{prefix}*",
                count=100,
            )
            keys.extend(batch)
            if cursor == 0:
                break

        total = len(keys)
        offset = (page - 1) * limit
        page_keys = keys[offset : offset + limit]

        sessions = []
        for key in page_keys:
            telegram_id = key.removeprefix(prefix)
            raw = await self._redis.get(key)
            roast_count = 0
            if raw:
                try:
                    payload = json.loads(raw)
                    if isinstance(payload, list):
                        roast_count = len(payload)
                except json.JSONDecodeError:
                    pass
            sessions.append({"telegram_id": telegram_id, "roast_count": roast_count})

        return sessions, total

    async def count_profile_roast_sessions(self) -> int:
        from config.profile_roast import profile_roast_settings

        prefix = profile_roast_settings.KEY_PREFIX
        count = 0
        cursor = 0
        while True:
            cursor, batch = await self._redis.scan(
                cursor=cursor,
                match=f"{prefix}*",
                count=100,
            )
            count += len(batch)
            if cursor == 0:
                break
        return count

    async def get_system_prompt(self) -> dict:
        return await self._prompt_store.get_admin_view()

    async def update_system_prompt(self, prompt: str) -> dict:
        await self._prompt_store.set_custom(prompt)
        return await self._prompt_store.get_admin_view()

    async def check_health(self) -> dict:
        redis_ok = False
        try:
            redis_ok = await self._redis.ping()
        except Exception as exc:
            logger.error("Redis health check failed: %s", exc)
        return {"ok": redis_ok, "db_ok": False, "redis_ok": redis_ok}
