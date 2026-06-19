import logging

from langchain_community.chat_message_histories import RedisChatMessageHistory
from langchain_core.messages import HumanMessage, SystemMessage
from langchain_core.runnables.history import RunnableWithMessageHistory
from langchain_openai import ChatOpenAI

from config.settings import settings
from utils.response import is_invalid_llm_response

logger = logging.getLogger(__name__)


class OpenAIManager:
    def __init__(self, session_id: str, llm: ChatOpenAI | None = None):
        self.session_id = session_id
        self.model = "gpt-3.5-turbo"
        self.temperature = 0.2
        self.max_tokens = 3000
        self.system_message = "Ты полезный AI-ассистент. Отвечай на вопросы пользователя."

        if llm is None:
            self.llm = ChatOpenAI(
                model=self.model,
                temperature=self.temperature,
                max_tokens=self.max_tokens,
                openai_api_key=settings.OPENAI_API_KEY,
                openai_api_base=settings.OPENAI_BASE_URL,
            )
        else:
            self.llm = llm

        def get_session_history(session_id: str) -> RedisChatMessageHistory:
            return RedisChatMessageHistory(
                session_id=session_id,
                url=settings.REDIS_URL,
            )

        self.conversation = RunnableWithMessageHistory(
            self.llm,
            get_session_history,
        )

    async def make_request(self, prompt: str) -> str | None:
        try:
            messages = [
                SystemMessage(content=self.system_message),
                HumanMessage(content=prompt),
            ]
            response = await self.conversation.ainvoke(
                messages,
                config={"configurable": {"session_id": self.session_id}},
            )
            content = response.content
            if is_invalid_llm_response(content):
                logger.error("OpenAI returned invalid response: %s", content)
                return None
            return content
        except Exception as exc:
            logger.error("OpenAI request failed: %s", exc)
            return None
