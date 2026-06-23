import logging

from config.throttle import throttle_settings
from llm.history.profile_roast_store import ProfileRoastStore, build_profile_roast_entry
from llm.profile_analyzer import ProfileData
from ratelimit.limiter import SlidingWindowLimiter

logger = logging.getLogger(__name__)

profile_limiter = SlidingWindowLimiter(
    throttle_settings.PROFILE_LIMIT,
    throttle_settings.PROFILE_WINDOW_SECONDS,
)


async def save_profile_roast(telegram_id: str, profile: ProfileData, response: str) -> None:
    store = ProfileRoastStore(telegram_id)
    if not await store.ping():
        return
    try:
        await store.append(build_profile_roast_entry(profile, response))
    except Exception as exc:
        logger.error(
            "Failed to save profile roast for telegram_id=%s: %s",
            telegram_id,
            exc,
        )
