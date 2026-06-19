import logging

from config.settings import settings
from llm.choices import ProviderChoiceGPT
from llm.errors import G4F_FALLBACK_ERROR_MESSAGE, LLMUserFacingError
from llm.managers.factory import PROVIDERS_FACTORY
from llm.managers.g4f_manager import G4FManager
from utils.response import is_invalid_llm_response

logger = logging.getLogger(__name__)


class AnyManager:
    def __init__(self, session_id: str):
        self.session_id = session_id

    def _create_manager(self, name: ProviderChoiceGPT, manager_class: type, is_fallback: bool):
        if manager_class is G4FManager:
            return manager_class(
                session_id=self.session_id,
                fallback_mode=is_fallback,
            )
        return manager_class(session_id=self.session_id)

    async def make_request(self, prompt: str) -> str | None:
        last_error: Exception | None = None
        providers = list(PROVIDERS_FACTORY.items())

        for index, (name, manager_class) in enumerate(providers):
            is_g4f_fallback = (
                not settings.DEBUG
                and name == ProviderChoiceGPT.G4F
                and index > 0
            )
            logger.info("Trying LLM provider: %s", name)
            try:
                manager = self._create_manager(name, manager_class, is_g4f_fallback)
                response = await manager.make_request(prompt)
                if response and not is_invalid_llm_response(response):
                    return response
            except LLMUserFacingError:
                raise
            except Exception as exc:
                last_error = exc
                logger.error("Provider %s failed: %s", name, exc)

        if last_error:
            logger.error("All LLM providers failed, last error: %s", last_error)
        raise LLMUserFacingError(G4F_FALLBACK_ERROR_MESSAGE)
