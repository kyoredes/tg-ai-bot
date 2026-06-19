import logging

from llm.g4f_setup import disable_g4f_media_writes

disable_g4f_media_writes()

from g4f.client import AsyncClient

from config.prompts import G4F_SYSTEM_PROMPT
from config.settings import settings
from llm.errors import G4F_FALLBACK_ERROR_MESSAGE, LLMUserFacingError
from llm.g4f_models import G4F_FALLBACK_MODELS, G4F_MODELS, G4FModelConfig
from llm.history.messages import to_openai_messages
from llm.history.store import ChatHistoryStore
from utils.response import is_invalid_llm_response

logger = logging.getLogger(__name__)


class G4FManager:
    def __init__(self, session_id: str, *, fallback_mode: bool = False):
        self.client = AsyncClient()
        self.fallback_mode = fallback_mode
        self.history = ChatHistoryStore(session_id)

    def _models_to_try(self) -> tuple[G4FModelConfig, ...]:
        if self.fallback_mode or settings.DEBUG:
            return G4F_FALLBACK_MODELS
        return G4F_MODELS

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
        await self.history.append(role, content)

    async def make_request(self, message: str) -> str | None:
        await self.history.ping()
        await self._add_message("user", message)

        try:
            messages = to_openai_messages(
                G4F_SYSTEM_PROMPT,
                await self.history.load(),
            )
            response = await self._get_response(messages)
            await self._add_message("assistant", response)
            return response
        except LLMUserFacingError:
            raise
        except Exception as exc:
            logger.error("G4F request failed: %s", exc)
            return None
