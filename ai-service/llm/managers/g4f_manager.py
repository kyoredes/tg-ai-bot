import json
import logging

from llm.g4f_setup import disable_g4f_media_writes

disable_g4f_media_writes()

from g4f.client import AsyncClient
from redis.asyncio import Redis
from redis.backoff import NoBackoff
from redis.retry import Retry

from config.settings import settings
from llm.errors import G4F_FALLBACK_ERROR_MESSAGE, LLMUserFacingError
from llm.g4f_models import G4F_FALLBACK_MODELS, G4F_MODELS, G4FModelConfig
from utils.response import is_invalid_llm_response
from utils.text import truncate_text_by_words

logger = logging.getLogger(__name__)


class G4FManager:
    def __init__(self, session_id: str, *, fallback_mode: bool = False):
        self.client = AsyncClient()
        self.fallback_mode = fallback_mode
        self.redis_client = Redis(
            host=settings.REDIS_HOST,
            port=settings.REDIS_PORT,
            db=settings.REDIS_DB,
            password=settings.REDIS_PASSWORD or None,
            decode_responses=True,
            socket_timeout=5,
            socket_connect_timeout=5,
            retry_on_timeout=False,
            retry=Retry(NoBackoff(), 0),
        )
        self.history_key = f"g4f_history:{session_id}"
        self.default_history = [
            {
                "role": "system",
                "content": (
                    "Ты полезный AI-ассистент. Отвечай на вопросы пользователя "
                    "только обычным текстом, без аудио, изображений и файлов."
                ),
            },
        ]
        self.redis_status = True

    def _models_to_try(self) -> tuple[G4FModelConfig, ...]:
        if self.fallback_mode or settings.DEBUG:
            return G4F_FALLBACK_MODELS
        return G4F_MODELS

    def _sanitize_history(self, history: list[dict]) -> list[dict]:
        cleaned = [self.default_history[0]]
        for message in history:
            if message.get("role") == "system":
                continue
            if is_invalid_llm_response(message.get("content")):
                continue
            cleaned.append(message)
        return cleaned if len(cleaned) > 1 else list(self.default_history)

    async def _load_history(self) -> list[dict]:
        if not self.redis_status:
            return list(self.default_history)
        try:
            history = await self.redis_client.get(self.history_key)
            if history:
                return self._sanitize_history(json.loads(history))
            await self._save_history(self.default_history)
            return list(self.default_history)
        except Exception as exc:
            logger.error("Error loading history: %s", exc)
            return list(self.default_history)

    async def _save_history(self, history: list[dict]) -> None:
        if not self.redis_status:
            return
        try:
            await self.redis_client.set(
                name=self.history_key,
                value=json.dumps(history or self.default_history),
                ex=60 * 60 * 24 * 7,
            )
        except Exception as exc:
            logger.error("Error saving history: %s", exc)

    async def _request_with_model(
        self,
        history: list[dict],
        model: G4FModelConfig,
    ) -> str | None:
        kwargs: dict = {
            "model": model.name,
            "messages": history,
            "web_search": False,
        }
        if model.provider is not None:
            kwargs["provider"] = model.provider

        response = await self.client.chat.completions.create(**kwargs)
        content = response.choices[0].message.content
        if not isinstance(content, str):
            logger.warning(
                "G4F model %s returned non-text response: %s",
                model.name,
                type(content).__name__,
            )
            return None
        if is_invalid_llm_response(content):
            logger.warning(
                "G4F model %s returned invalid response: %s",
                model.name,
                content,
            )
            return None
        return content

    async def _get_response(self, history: list[dict]) -> str | None:
        models = self._models_to_try()
        last_error: Exception | None = None

        for index, model in enumerate(models):
            try:
                if index > 0:
                    logger.info(
                        "Switching to next G4F model: %s (%d/%d)",
                        model,
                        index + 1,
                        len(models),
                    )
                else:
                    logger.info("G4F request via model: %s", model)

                content = await self._request_with_model(history, model)
                if content:
                    logger.info("G4F success with model: %s", model.name)
                    return content
            except Exception as exc:
                last_error = exc
                logger.warning(
                    "G4F model %s failed: %s",
                    model.name,
                    exc,
                )

        if last_error:
            logger.error(
                "All G4F models failed after %d attempts, last error: %s",
                len(models),
                last_error,
            )
        else:
            logger.error(
                "All G4F models returned invalid responses after %d attempts",
                len(models),
            )
        raise LLMUserFacingError(G4F_FALLBACK_ERROR_MESSAGE)

    async def _add_message(self, role: str, content: str) -> None:
        if is_invalid_llm_response(content):
            return
        message = await truncate_text_by_words(content)
        history = await self._load_history()
        history.append({"role": role, "content": message})
        await self._save_history(history)

    async def _check_redis(self) -> bool:
        try:
            await self.redis_client.ping()
            return True
        except Exception as exc:
            logger.error("Redis ping failed: %s", exc)
            return False

    async def make_request(self, message: str) -> str | None:
        self.redis_status = await self._check_redis()
        await self._add_message("user", message)

        try:
            history = await self._load_history()
            response = await self._get_response(history)
            await self._add_message("assistant", response)
            return response
        except LLMUserFacingError:
            raise
        except Exception as exc:
            logger.error("G4F request failed: %s", exc)
            return None
