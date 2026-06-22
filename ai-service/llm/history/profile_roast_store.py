import json
import logging
import time

from pydantic import BaseModel
from redis.asyncio import Redis
from redis.backoff import NoBackoff
from redis.retry import Retry

from config.profile_roast import profile_roast_settings
from config.settings import settings
from utils.response import is_invalid_llm_response
from utils.text import truncate_to_word_limit

logger = logging.getLogger(__name__)


class ProfileRoastEntry(BaseModel):
    created_at: int
    first_name: str
    last_name: str = ""
    username: str = ""
    bio: str = ""
    is_premium: bool = False
    language_code: str = ""
    has_photo: bool = False
    response: str


class ProfileRoastStore:
    def __init__(self, telegram_id: str, redis_client: Redis | None = None):
        self.telegram_id = telegram_id
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
        return f"{profile_roast_settings.KEY_PREFIX}{self.telegram_id}"

    async def ping(self) -> bool:
        try:
            await self._redis.ping()
            self._available = True
            return True
        except Exception as exc:
            logger.error("Redis ping failed for profile roast store: %s", exc)
            self._available = False
            return False

    async def load(self) -> list[ProfileRoastEntry]:
        if not self._available:
            return []

        raw = await self._redis.get(self.key)
        if not raw:
            return []

        try:
            payload = json.loads(raw)
        except json.JSONDecodeError:
            logger.error("Invalid profile roast JSON for telegram_id=%s", self.telegram_id)
            return []

        if not isinstance(payload, list):
            return []

        entries: list[ProfileRoastEntry] = []
        for item in payload:
            if not isinstance(item, dict):
                continue
            try:
                entries.append(ProfileRoastEntry.model_validate(item))
            except Exception:
                continue
        return entries

    async def append(self, entry: ProfileRoastEntry) -> None:
        if not self._available or is_invalid_llm_response(entry.response):
            return

        entries = await self.load()
        entries.append(entry)
        await self._save(entries)

    async def clear(self) -> None:
        if not self._available:
            return
        try:
            await self._redis.delete(self.key)
        except Exception as exc:
            logger.error("Error clearing profile roast history: %s", exc)

    async def _save(self, entries: list[ProfileRoastEntry]) -> None:
        if not self._available:
            return

        cleaned = self._trim(entries)
        try:
            await self._redis.set(
                name=self.key,
                value=json.dumps(
                    [entry.model_dump() for entry in cleaned],
                    ensure_ascii=False,
                ),
                ex=profile_roast_settings.TTL_SECONDS,
            )
        except Exception as exc:
            logger.error("Error saving profile roast history: %s", exc)

    def _trim(self, entries: list[ProfileRoastEntry]) -> list[ProfileRoastEntry]:
        trimmed = [
            ProfileRoastEntry(
                created_at=entry.created_at,
                first_name=entry.first_name,
                last_name=entry.last_name,
                username=entry.username,
                bio=entry.bio,
                is_premium=entry.is_premium,
                language_code=entry.language_code,
                has_photo=entry.has_photo,
                response=truncate_to_word_limit(entry.response, 800),
            )
            for entry in entries
        ]

        max_entries = profile_roast_settings.MAX_ENTRIES
        if max_entries > 0 and len(trimmed) > max_entries:
            trimmed = trimmed[-max_entries:]
        return trimmed


def build_profile_roast_entry(profile, response: str) -> ProfileRoastEntry:
    return ProfileRoastEntry(
        created_at=int(time.time()),
        first_name=profile.first_name,
        last_name=profile.last_name,
        username=profile.username,
        bio=profile.bio,
        is_premium=profile.is_premium,
        language_code=profile.language_code,
        has_photo=bool(profile.photo_base64),
        response=response,
    )
