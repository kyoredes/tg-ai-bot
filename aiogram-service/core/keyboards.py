from aiogram.types import (
    InlineKeyboardMarkup,
    InlineKeyboardButton,
    ReplyKeyboardMarkup,
)
from aiogram.utils.keyboard import ReplyKeyboardBuilder


def get_start_keyboard():
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            [InlineKeyboardButton(text="🤖 К нейросети", callback_data="neuro")],
            [InlineKeyboardButton(text="🧐 О тарифе", callback_data="subscription")],
        ]
    )
    return keyboard


def to_main_menu_keyboard() -> ReplyKeyboardMarkup:
    builder = ReplyKeyboardBuilder()

    builder.button(text="👈 Назад")

    return builder.as_markup(resize_keyboard=True)

def to_main_keyboard_inline() -> InlineKeyboardMarkup:
    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            [InlineKeyboardButton(text="👈 Назад", callback_data="main_menu")],
        ]
    )
    return keyboard
