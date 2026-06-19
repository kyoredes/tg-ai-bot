from aiogram import Router, F, types
from aiogram.filters import CommandStart
from aiogram.types import Message
from core.keyboards import get_profile_keyboard, get_start_keyboard
from users.answers import get_profile_info_answer
from users.manager import UserManager

users_router = Router(name="users")


@users_router.message(CommandStart())
async def start(message: Message):
    user_manager = UserManager()
    client = await user_manager.start(str(message.from_user.id))
    if client is None:
        await message.answer("Не удалось создать клиента")
        return

    await message.answer("Привет! Выбери действие:", reply_markup=get_start_keyboard())


@users_router.callback_query(F.data == "profile")
async def get_profile(callback: types.CallbackQuery):
    user_manager = UserManager()
    client = await user_manager.get_profile(str(callback.from_user.id))
    if client is None:
        await callback.message.answer("Не удалось загрузить профиль")
        await callback.answer()
        return

    answer = await get_profile_info_answer(client)
    await callback.message.answer(
        answer,
        parse_mode="Markdown",
        reply_markup=get_profile_keyboard(),
    )
    await callback.answer()
