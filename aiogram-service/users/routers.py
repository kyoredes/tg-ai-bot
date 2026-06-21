from aiogram import Router, F, types
from aiogram.filters import CommandStart
from aiogram.fsm.context import FSMContext
from aiogram.types import Message
from core.keyboards import get_back_keyboard, get_neuro_keyboard, get_start_keyboard
from core.typing_indicator import show_typing
from users.answers import (
    CHAT_ERROR_ANSWER,
    CLEAR_CONTEXT_ERROR_ANSWER,
    CLEAR_CONTEXT_OK_ANSWER,
    get_profile_info_answer,
    get_subscription_info_answer,
)
from users.manager import UserManager
from users.states import NeuroStates

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
        reply_markup=get_back_keyboard(),
    )
    await callback.answer()


@users_router.callback_query(F.data == "subscription")
async def get_subscription(callback: types.CallbackQuery):
    user_manager = UserManager()
    subscription = await user_manager.get_subscription(str(callback.from_user.id))
    if subscription is None:
        await callback.message.answer("Не удалось загрузить информацию о тарифе")
        await callback.answer()
        return

    answer = await get_subscription_info_answer(subscription)
    await callback.message.answer(
        answer,
        parse_mode="Markdown",
        reply_markup=get_back_keyboard(),
    )
    await callback.answer()


@users_router.callback_query(F.data == "neuro")
async def neuro_start(callback: types.CallbackQuery, state: FSMContext):
    await state.set_state(NeuroStates.waiting_prompt)
    await callback.message.answer(
        "Напиши свой вопрос нейросети:",
        reply_markup=get_neuro_keyboard(),
    )
    await callback.answer()


@users_router.message(NeuroStates.waiting_prompt)
async def neuro_chat(message: Message, state: FSMContext):
    try:
        user_manager = UserManager()
        async with show_typing(message):
            chat = await user_manager.chat(str(message.from_user.id), message.text or "")
        if chat is None or not chat.response:
            await message.answer(CHAT_ERROR_ANSWER, reply_markup=get_neuro_keyboard())
            return

        await message.answer(chat.response, reply_markup=get_neuro_keyboard())
    except Exception:
        await message.answer(CHAT_ERROR_ANSWER, reply_markup=get_neuro_keyboard())


@users_router.callback_query(F.data == "clear_context")
async def clear_context(callback: types.CallbackQuery, state: FSMContext):
    user_manager = UserManager()
    cleared = await user_manager.clear_chat(str(callback.from_user.id))
    if not cleared:
        await callback.message.answer(
            CLEAR_CONTEXT_ERROR_ANSWER,
            reply_markup=get_neuro_keyboard(),
        )
        await callback.answer()
        return

    await state.set_state(NeuroStates.waiting_prompt)
    await callback.message.answer(
        CLEAR_CONTEXT_OK_ANSWER,
        reply_markup=get_neuro_keyboard(),
    )
    await callback.answer()
