import base64
import logging

from aiogram import Bot
from aiogram.types import User

from users.schemas import ProfileSnapshot

logger = logging.getLogger(__name__)


async def collect_profile_snapshot(bot: Bot, user: User) -> ProfileSnapshot:
    first_name = user.first_name or ""
    last_name = user.last_name or ""
    username = user.username or ""
    is_premium = bool(user.is_premium)
    language_code = user.language_code or ""

    bio = ""
    try:
        chat = await bot.get_chat(user.id)
        bio = chat.bio or ""
    except Exception as exc:
        logger.debug("Failed to load bio for user %s: %s", user.id, exc)

    photo_base64: str | None = None
    has_photo = False
    try:
        photos = await bot.get_user_profile_photos(user.id, limit=1)
        if photos.total_count > 0 and photos.photos:
            largest = photos.photos[0][-1]
            file = await bot.get_file(largest.file_id)
            if file.file_path:
                downloaded = await bot.download(file)
                if hasattr(downloaded, "read"):
                    data = downloaded.read()
                else:
                    data = downloaded
                if data:
                    photo_base64 = base64.b64encode(data).decode("ascii")
                    has_photo = True
    except Exception as exc:
        logger.debug("Failed to load profile photo for user %s: %s", user.id, exc)

    return ProfileSnapshot(
        first_name=first_name,
        last_name=last_name,
        username=username,
        bio=bio,
        is_premium=is_premium,
        language_code=language_code,
        photo_base64=photo_base64,
        has_photo=has_photo,
    )
