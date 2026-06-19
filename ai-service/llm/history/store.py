import json
import logging

from langchain_core.messages import messages_from_dict
from redis.asyncio import Redis
from redis.backoff import NoBackoff
from redis.retry import Retry

from config.history import chat_history_settings
from config.settings import settings
from llm.history.messages import ChatMessage
from utils.response import is_invalid_llm_response
from utils.text import count_words, truncate_to_word_limit

logger = logging.getLogger(__name__)


class ChatHistoryStore:
    def __init__(self, session_id: str, redis_client: Redis | None = None):
        self.session_id = session_id
        self._redis = redis_client or Redis.from_url(
            settings.REDIS_URL,
            decode_responses=True,
            socket_timeout=5,
            socket_connect_timeout=5,
            retry_on_timeout=False,
            retry=Retry(NoBackoff(), 0),
        )
        self._available = True

    @property
    def key(self) -> str:
        return f"{chat_history_settings.KEY_PREFIX}{self.session_id}"

    async def ping(self) -> bool:
        try:
            await self._redis.ping()
            self._available = True
            return True
        except Exception as exc:
            logger.error("Redis ping failed: %s", exc)
            self._available = False
            return False

    @property
    def available(self) -> bool:
        return self._available

    async def load(self) -> list[ChatMessage]:
        if not self._available:
            return []

        raw = await self._redis.get(self.key)
        if raw:
            return self._trim(self._parse_messages(raw))

        migrated = await self._migrate_legacy()
        if migrated:
            await self._save(migrated)
            return migrated

        return []

    async def save(self, messages: list[ChatMessage]) -> None:
        await self._save(messages)

    async def append(self, role: str, content: str) -> None:
        if not self._available or is_invalid_llm_response(content):
            return

        messages = await self.load()
        messages.append(ChatMessage(role=role, content=content))
        await self._save(messages)

    async def _save(self, messages: list[ChatMessage]) -> None:
        if not self._available:
            return

        cleaned = self._sanitize(messages)
        try:
            await self._redis.set(
                name=self.key,
                value=json.dumps([message.model_dump() for message in cleaned]),
                ex=chat_history_settings.TTL_SECONDS,
            )
        except Exception as exc:
            logger.error("Error saving chat history: %s", exc)

    def _sanitize(self, messages: list[ChatMessage]) -> list[ChatMessage]:
        cleaned = [
            message
            for message in messages
            if message.role in ("user", "assistant")
            and not is_invalid_llm_response(message.content)
        ]
        return self._trim(cleaned)

    def _trim(self, messages: list[ChatMessage]) -> list[ChatMessage]:
        trimmed = [
            ChatMessage(
                role=message.role,
                content=truncate_to_word_limit(
                    message.content,
                    chat_history_settings.MAX_WORDS_PER_MESSAGE,
                ),
            )
            for message in messages
        ]

        max_messages = chat_history_settings.MAX_MESSAGES
        if max_messages > 0 and len(trimmed) > max_messages:
            trimmed = trimmed[-max_messages:]

        max_total_words = chat_history_settings.MAX_TOTAL_WORDS
        if max_total_words > 0:
            while trimmed and self._total_words(trimmed) > max_total_words:
                trimmed = trimmed[1:]

        return trimmed

    @staticmethod
    def _total_words(messages: list[ChatMessage]) -> int:
        return sum(count_words(message.content) for message in messages)

    def _parse_messages(self, raw: str) -> list[ChatMessage]:
        try:
            payload = json.loads(raw)
        except json.JSONDecodeError:
            logger.error("Invalid chat history JSON for session %s", self.session_id)
            return []

        if not isinstance(payload, list):
            return []

        messages: list[ChatMessage] = []
        for item in payload:
            if not isinstance(item, dict):
                continue
            role = item.get("role")
            content = item.get("content")
            if role not in ("user", "assistant") or not isinstance(content, str):
                continue
            if is_invalid_llm_response(content):
                continue
            messages.append(ChatMessage(role=role, content=content))
        return messages

    async def _migrate_legacy(self) -> list[ChatMessage]:
        g4f_key = f"{chat_history_settings.LEGACY_G4F_PREFIX}{self.session_id}"
        g4f_raw = await self._redis.get(g4f_key)
        if g4f_raw:
            messages = self._parse_legacy_g4f(g4f_raw)
            if messages:
                logger.info("Migrated G4F history for session %s", self.session_id)
                await self._redis.delete(g4f_key)
                return messages

        langchain_key = f"{chat_history_settings.LEGACY_LANGCHAIN_PREFIX}{self.session_id}"
        langchain_raw = await self._redis.lrange(langchain_key, 0, -1)
        if langchain_raw:
            messages = self._parse_legacy_langchain(langchain_raw)
            if messages:
                logger.info("Migrated LangChain history for session %s", self.session_id)
                await self._redis.delete(langchain_key)
                return messages

        return []

    def _parse_legacy_g4f(self, raw: str) -> list[ChatMessage]:
        try:
            payload = json.loads(raw)
        except json.JSONDecodeError:
            return []

        if not isinstance(payload, list):
            return []

        messages: list[ChatMessage] = []
        for item in payload:
            if not isinstance(item, dict):
                continue
            role = item.get("role")
            content = item.get("content")
            if role == "system":
                continue
            if role not in ("user", "assistant") or not isinstance(content, str):
                continue
            if is_invalid_llm_response(content):
                continue
            messages.append(ChatMessage(role=role, content=content))
        return messages

    def _parse_legacy_langchain(self, raw_items: list[str]) -> list[ChatMessage]:
        try:
            items = [json.loads(item) for item in raw_items[::-1]]
            langchain_messages = messages_from_dict(items)
        except (json.JSONDecodeError, TypeError, ValueError) as exc:
            logger.error("Failed to parse LangChain history: %s", exc)
            return []

        messages: list[ChatMessage] = []
        for message in langchain_messages:
            content = message.content
            if not isinstance(content, str) or is_invalid_llm_response(content):
                continue
            if message.type == "human":
                messages.append(ChatMessage(role="user", content=content))
            elif message.type == "ai":
                messages.append(ChatMessage(role="assistant", content=content))
        return messages
