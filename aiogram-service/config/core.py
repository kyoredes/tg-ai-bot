from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    BOT_TOKEN: str
    GATEWAY_HOST: str
    GATEWAY_PORT: str
    COMMON_PUB_KEY: str
    LOG_LEVEL: str = "INFO"
    HTTP_TIMEOUT: float = 120.0
    KAFKA_BROKERS: str = "localhost:9092"
    KAFKA_GROUP_ID: str = "aiogram-profile-results"
    THROTTLE_ENABLED: bool = True
    THROTTLE_LIMIT: int = 5
    THROTTLE_WINDOW: float = 5.0
    THROTTLE_BAN_SECONDS: float = 100.0

    class Config:
        env_file = ".env"

    def kafka_broker_list(self) -> list[str]:
        return [broker.strip() for broker in self.KAFKA_BROKERS.split(",") if broker.strip()]

settings = Settings()