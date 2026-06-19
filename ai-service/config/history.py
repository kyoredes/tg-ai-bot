from pydantic_settings import BaseSettings, SettingsConfigDict


class ChatHistorySettings(BaseSettings):
    KEY_PREFIX: str = "chat_history:"
    TTL_SECONDS: int = 60 * 60 * 24 * 7
    LEGACY_G4F_PREFIX: str = "g4f_history:"
    LEGACY_LANGCHAIN_PREFIX: str = "message_store:"
    MAX_MESSAGES: int = 20
    MAX_WORDS_PER_MESSAGE: int = 400
    MAX_TOTAL_WORDS: int = 3000

    model_config = SettingsConfigDict(
        env_prefix="CHAT_HISTORY_",
        env_file=".env",
        extra="ignore",
    )


chat_history_settings = ChatHistorySettings()
