from pydantic_settings import BaseSettings, SettingsConfigDict


class LiteLLMSettings(BaseSettings):
    MODEL: str = "gpt-3.5-turbo"
    TEMPERATURE: float = 0.2
    MAX_TOKENS: int = 3000
    API_KEY: str = ""
    API_BASE: str = ""
    CUSTOM_PROVIDER: str = ""
    REQUEST_TIMEOUT: float = 120.0
    MAX_RETRIES: int = 2

    model_config = SettingsConfigDict(
        env_prefix="LITELLM_",
        env_file=".env",
        extra="ignore",
    )

    def resolved_api_key(self, fallback: str = "") -> str:
        return self.API_KEY or fallback

    def resolved_api_base(self, fallback: str = "") -> str | None:
        base = self.API_BASE or fallback
        return base or None


litellm_settings = LiteLLMSettings()
