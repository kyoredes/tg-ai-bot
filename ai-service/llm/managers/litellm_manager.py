import logging

from langchain_litellm import ChatLiteLLM

from config.llm import litellm_settings
from config.prompts import ASSISTANT_SYSTEM_PROMPT
from config.settings import settings
from llm.history.messages import to_langchain_messages
from llm.history.store import ChatHistoryStore
from utils.response import is_invalid_llm_response

logger = logging.getLogger(__name__)


class LiteLLMManager:
    def __init__(self, session_id: str, llm: ChatLiteLLM | None = None):
        self.session_id = session_id
        self.llm = llm or self._build_llm()
        self.history = ChatHistoryStore(session_id)

    def _build_llm(self) -> ChatLiteLLM:
        kwargs: dict = {
            "model": litellm_settings.MODEL,
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

    async def make_request(self, prompt: str) -> str | None:
        await self.history.ping()
        await self.history.append("user", prompt)

        try:
            messages = to_langchain_messages(
                ASSISTANT_SYSTEM_PROMPT,
                await self.history.load(),
            )
            response = await self.llm.ainvoke(messages)
            content = response.content
            if not isinstance(content, str) or is_invalid_llm_response(content):
                logger.error("LiteLLM returned invalid response: %s", content)
                return None

            await self.history.append("assistant", content)
            return content
        except Exception as exc:
            logger.error("LiteLLM request failed: %s", exc)
            return None
