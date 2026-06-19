from aiogram.types import (
    InlineKeyboardMarkup,
    InlineKeyboardButton,
)


def get_start_keyboard():
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [InlineKeyboardButton(text="🤖 К нейросети", callback_data="neuro")],
            [InlineKeyboardButton(text="🧐 О тарифе", callback_data="subscription")],
            [InlineKeyboardButton(text="👤Профиль", callback_data="profile")],
        ]
    )


def get_profile_keyboard():
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [InlineKeyboardButton(text="👈 Назад", callback_data="main_menu")],
        ]
    )
