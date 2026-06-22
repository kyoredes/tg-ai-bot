from pydantic_settings import BaseSettings, SettingsConfigDict

TOPIC_PROFILE_ANALYZE_REQUESTS = "profile.analyze.requests"
TOPIC_PROFILE_ANALYZE_RESULTS = "profile.analyze.results"


class KafkaSettings(BaseSettings):
    BROKERS: str = "localhost:9092"
    GROUP_ID: str = "ai-service-profile-worker"

    model_config = SettingsConfigDict(
        env_prefix="KAFKA_",
        env_file=".env",
        extra="ignore",
    )

    def broker_list(self) -> list[str]:
        return [broker.strip() for broker in self.BROKERS.split(",") if broker.strip()]


kafka_settings = KafkaSettings()
