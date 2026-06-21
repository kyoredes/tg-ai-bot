from pydantic_settings import BaseSettings, SettingsConfigDict


class ThrottleSettings(BaseSettings):
    ENABLED: bool = True
    LIMIT: int = 300
    WINDOW_SECONDS: float = 60.0
    CHAT_LIMIT: int = 10
    CHAT_WINDOW_SECONDS: float = 60.0

    model_config = SettingsConfigDict(
        env_prefix="THROTTLE_",
        env_file=".env",
        extra="ignore",
    )


throttle_settings = ThrottleSettings()
