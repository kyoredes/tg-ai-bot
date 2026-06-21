import time
from collections import defaultdict, deque
from typing import Any

from aiogram import BaseMiddleware
from aiogram.dispatcher.event.bases import SkipHandler
from aiogram.types import CallbackQuery, Message, TelegramObject


class ThrottlingMiddleware(BaseMiddleware):
    def __init__(
        self,
        limit_count: int = 5,
        limit_seconds: float = 5.0,
        ban_seconds: float = 100.0,
    ):
        self.limit_count = limit_count
        self.limit_seconds = limit_seconds
        self.ban_seconds = ban_seconds

        self.user_message_times: dict[int, deque[float]] = defaultdict(deque)
        self.banned_users: dict[int, float] = {}

    def _user_id(self, event: TelegramObject) -> int | None:
        user = getattr(event, "from_user", None)
        if user is None:
            return None
        return user.id

    async def __call__(
        self,
        handler,
        event: TelegramObject,
        data: dict[str, Any],
    ):
        user_id = self._user_id(event)
        if user_id is None:
            return await handler(event, data)

        now = time.monotonic()

        ban_until = self.banned_users.get(user_id)
        if ban_until and now < ban_until:
            await self._notify(event, "🚫 Ты временно забанен за спам. Подожди немного.")
            raise SkipHandler()

        times = self.user_message_times[user_id]
        times.append(now)

        while times and now - times[0] > self.limit_seconds:
            times.popleft()

        if len(times) > self.limit_count:
            self.banned_users[user_id] = now + self.ban_seconds
            await self._notify(event, "🚫 Слишком много сообщений! Ты временно забанен.")
            raise SkipHandler()

        return await handler(event, data)

    @staticmethod
    async def _notify(event: TelegramObject, text: str) -> None:
        if isinstance(event, Message):
            await event.answer(text)
        elif isinstance(event, CallbackQuery):
            await event.answer(text, show_alert=True)
