import logging

from llm.choices import ProviderChoiceGPT
from llm.managers.any_manager import AnyManager
from llm.managers.factory import PROVIDERS_FACTORY

logger = logging.getLogger(__name__)


class LLMManager:
    def __init__(
        self,
        session_id: str,
        provider: ProviderChoiceGPT = ProviderChoiceGPT.ANY,
    ):
        self.provider = provider
        self.session_id = session_id

    async def make_request(self, prompt: str) -> str | None:
        if self.provider == ProviderChoiceGPT.ANY:
            any_manager = AnyManager(session_id=self.session_id)
            return await any_manager.make_request(prompt)

        manager_class = PROVIDERS_FACTORY.get(self.provider)
        if manager_class is None:
            raise ValueError(f"Invalid provider: {self.provider}")

        manager = manager_class(session_id=self.session_id)
        return await manager.make_request(prompt)
