import asyncio
import logging
from config.core import settings
from aiogram import Bot, Dispatcher
from middleware.throttling import ThrottlingMiddleware
from core.routers import main_router
from users.routers import users_router

LOG_LEVEL = settings.LOG_LEVEL
logger = logging.getLogger(__name__)

logging.basicConfig(level=LOG_LEVEL)

bot = Bot(token=settings.BOT_TOKEN)
dp = Dispatcher()
dp.message.middleware(ThrottlingMiddleware())
dp.include_router(main_router)
dp.include_router(users_router)


async def main():
    await dp.start_polling(bot)


if __name__ == "__main__":
    print("Starting bot...")
    asyncio.run(main())

