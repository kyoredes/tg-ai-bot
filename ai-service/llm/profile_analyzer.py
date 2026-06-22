import asyncio
import base64
import json
import logging

from langchain_core.messages import HumanMessage, SystemMessage
from langchain_litellm import ChatLiteLLM

from config.llm import litellm_settings
from config.prompts import G4F_INSTRUCTION_ACK, PROFILE_ROAST_SYSTEM_PROMPT
from config.settings import settings
from llm.errors import G4F_FALLBACK_ERROR_MESSAGE, LLMUserFacingError
from llm.g4f_models import G4F_FALLBACK_MODELS, G4F_MODELS, G4FModelConfig
from llm.g4f_setup import disable_g4f_media_writes
from utils.response import is_invalid_llm_response

disable_g4f_media_writes()

from g4f.client import AsyncClient  # noqa: E402

logger = logging.getLogger(__name__)

G4F_MODEL_TIMEOUT_SECONDS = 30.0
DEV_G4F_MODEL_TIMEOUT_SECONDS = 90.0
DEV_PROFILE_ANALYZE_MAX_ATTEMPTS = 3


def _g4f_model_timeout_seconds() -> float:
    return DEV_G4F_MODEL_TIMEOUT_SECONDS if settings.DEBUG else G4F_MODEL_TIMEOUT_SECONDS


class ProfileData:
    def __init__(
        self,
        *,
        first_name: str,
        last_name: str = "",
        username: str = "",
        bio: str = "",
        is_premium: bool = False,
        language_code: str = "",
        photo_base64: str = "",
    ):
        self.first_name = first_name
        self.last_name = last_name
        self.username = username
        self.bio = bio
        self.is_premium = is_premium
        self.language_code = language_code
        self.photo_base64 = photo_base64

    def to_prompt_text(self, *, include_photo_note: bool = True) -> str:
        payload = {
            "first_name": self.first_name or None,
            "last_name": self.last_name or None,
            "username": f"@{self.username}" if self.username else None,
            "bio": self.bio or None,
            "is_premium": self.is_premium,
            "language_code": self.language_code or None,
            "has_profile_photo": bool(self.photo_base64),
        }
        if include_photo_note and not self.photo_base64:
            payload["photo_note"] = "фото профиля недоступно"
        return (
            "Проанализируй Telegram-профиль пользователя по этим данным:\n"
            f"{json.dumps(payload, ensure_ascii=False, indent=2)}"
        )


class ProfileAnalyzer:
    def _decode_photo(self, profile: ProfileData) -> bytes | None:
        if not profile.photo_base64:
            return None
        try:
            return base64.b64decode(profile.photo_base64)
        except Exception as exc:
            logger.warning("Invalid profile photo base64: %s", exc)
            return None

    def _build_litellm(self) -> ChatLiteLLM:
        kwargs: dict = {
            "model": litellm_settings.PROFILE_MODEL,
            "temperature": litellm_settings.TEMPERATURE,
            "max_tokens": litellm_settings.MAX_TOKENS,
            "api_key": litellm_settings.resolved_api_key(settings.OPENAI_API_KEY),
            "max_retries": litellm_settings.MAX_RETRIES,
            "request_timeout": litellm_settings.REQUEST_TIMEOUT,
        }
        api_base = litellm_settings.resolved_api_base(settings.OPENAI_BASE_URL)
        if api_base:
            kwargs["api_base"] = api_base
        if litellm_settings.CUSTOM_PROVIDER:
            kwargs["custom_llm_provider"] = litellm_settings.CUSTOM_PROVIDER
        return ChatLiteLLM(**kwargs)

    def _build_human_message(self, profile: ProfileData) -> HumanMessage:
        text = profile.to_prompt_text(include_photo_note=not profile.photo_base64)
        if profile.photo_base64:
            content: list[dict] = [
                {"type": "text", "text": text},
                {
                    "type": "image_url",
                    "image_url": {
                        "url": f"data:image/jpeg;base64,{profile.photo_base64}",
                    },
                },
            ]
            return HumanMessage(content=content)
        return HumanMessage(content=text)

    async def _request_litellm(self, profile: ProfileData) -> str | None:
        llm = self._build_litellm()
        messages = [
            SystemMessage(content=PROFILE_ROAST_SYSTEM_PROMPT),
            self._build_human_message(profile),
        ]
        response = await llm.ainvoke(messages)
        content = response.content
        if not isinstance(content, str) or is_invalid_llm_response(content):
            logger.error("Profile LiteLLM returned invalid response: %s", content)
            return None
        return content

    async def _request_g4f(
        self,
        profile: ProfileData,
        *,
        image_bytes: bytes | None = None,
    ) -> str | None:
        client = AsyncClient()
        models: tuple[G4FModelConfig, ...] = (
            G4F_FALLBACK_MODELS if settings.DEBUG else G4F_MODELS
        )
        user_prompt = profile.to_prompt_text(include_photo_note=image_bytes is None)
        if image_bytes is not None:
            user_prompt += "\n\nК запросу приложено фото профиля — учти его при анализе."
        messages = [
            {"role": "user", "content": PROFILE_ROAST_SYSTEM_PROMPT},
            {"role": "assistant", "content": G4F_INSTRUCTION_ACK},
            {"role": "user", "content": user_prompt},
        ]

        timeout_seconds = _g4f_model_timeout_seconds()
        last_error: Exception | None = None
        for index, model in enumerate(models):
            kwargs: dict = {
                "model": model.name,
                "messages": messages,
                "web_search": False,
            }
            if model.provider is not None:
                kwargs["provider"] = model.provider
            if image_bytes is not None:
                kwargs["image"] = image_bytes
                kwargs["image_name"] = "profile.jpg"
            try:
                if index > 0:
                    logger.info("Profile G4F switching to model: %s", model.name)
                response = await asyncio.wait_for(
                    client.chat.completions.create(**kwargs),
                    timeout=timeout_seconds,
                )
                content = response.choices[0].message.content
                if isinstance(content, str) and not is_invalid_llm_response(content):
                    return content
            except TimeoutError:
                last_error = TimeoutError(f"model {model.name} timed out")
                logger.warning(
                    "Profile G4F model %s timed out after %ss",
                    model.name,
                    timeout_seconds,
                )
            except Exception as exc:
                last_error = exc
                logger.warning("Profile G4F model %s failed: %s", model.name, exc)

        if last_error:
            logger.error("Profile G4F failed: %s", last_error)
        return None

    async def _analyze_via_g4f(
        self,
        profile: ProfileData,
        *,
        image_bytes: bytes | None,
    ) -> str | None:
        response = await self._request_g4f(profile, image_bytes=image_bytes)
        if response:
            return response

        if image_bytes is not None:
            logger.info("Profile G4F vision failed, retrying text-only")
            return await self._request_g4f(profile, image_bytes=None)
        return None

    async def analyze(self, profile: ProfileData) -> str:
        image_bytes = self._decode_photo(profile)

        if not settings.DEBUG:
            try:
                response = await self._request_litellm(profile)
                if response:
                    return response
            except Exception as exc:
                logger.warning("Profile LiteLLM failed, trying G4F: %s", exc)

        max_attempts = (
            DEV_PROFILE_ANALYZE_MAX_ATTEMPTS if settings.DEBUG else 1
        )
        for attempt in range(1, max_attempts + 1):
            if attempt > 1:
                logger.info(
                    "Profile G4F dev retry attempt %d/%d",
                    attempt,
                    max_attempts,
                )
            response = await self._analyze_via_g4f(profile, image_bytes=image_bytes)
            if response:
                return response

        raise LLMUserFacingError(G4F_FALLBACK_ERROR_MESSAGE)
