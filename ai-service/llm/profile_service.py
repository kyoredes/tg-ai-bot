import logging

from config.prompts import LLM_FALLBACK_ERROR_MESSAGE, PROFILE_RATE_LIMIT_MESSAGE
from config.throttle import throttle_settings
from llm.profile_persistence import profile_limiter, save_profile_roast
from llm.errors import LLMUserFacingError
from llm.profile_analyzer import ProfileAnalyzer, ProfileData
from utils.response import is_invalid_llm_response

logger = logging.getLogger(__name__)


async def run_profile_analysis(telegram_id: str, profile: ProfileData) -> str:
    if throttle_settings.ENABLED and not profile_limiter.allow(f"profile:{telegram_id}"):
        return PROFILE_RATE_LIMIT_MESSAGE

    try:
        analyzer = ProfileAnalyzer()
        response = await analyzer.analyze(profile)
    except LLMUserFacingError as exc:
        logger.warning(
            "User-facing profile analysis error for telegram_id=%s: %s",
            telegram_id,
            exc.message,
        )
        return exc.message
    except Exception:
        logger.exception("Profile analysis failed for telegram_id=%s", telegram_id)
        return LLM_FALLBACK_ERROR_MESSAGE

    if is_invalid_llm_response(response):
        logger.error(
            "Profile analysis returned invalid response for telegram_id=%s",
            telegram_id,
        )
        return LLM_FALLBACK_ERROR_MESSAGE

    await save_profile_roast(telegram_id, profile, response)
    return response
