from pydantic_settings import BaseSettings, SettingsConfigDict


class ProfileRoastSettings(BaseSettings):
    KEY_PREFIX: str = "profile_roast:"
    TTL_SECONDS: int = 60 * 60 * 24 * 7
    MAX_ENTRIES: int = 20

    model_config = SettingsConfigDict(
        env_prefix="PROFILE_ROAST_",
        env_file=".env",
        extra="ignore",
    )


profile_roast_settings = ProfileRoastSettings()
