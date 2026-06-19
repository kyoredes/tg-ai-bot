from aiogram import Router
from aiogram.filters import CommandStart
from aiogram.types import Message, InlineKeyboardMarkup
from core.keyboards import get_start_keyboard, to_main_menu_keyboard
from users.manager import UserManager
import logging
from aiogram import F, types
users_router = Router(name="users")

logger = logging.getLogger(__name__)


@users_router.message(CommandStart())
async def start(message: Message):
    user_manager = UserManager()
    client = await user_manager.start(str(message.from_user.id))
    if client is None:
        logger.error(f"Не удалось создать клиента {message.from_user.id}")
        await message.answer("Не удалось создать клиента")
        return

    await message.answer("Привет! Выбери действие:", reply_markup=get_start_keyboard())

