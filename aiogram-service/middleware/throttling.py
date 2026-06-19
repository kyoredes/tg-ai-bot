import time
from collections import defaultdict, deque
from aiogram import BaseMiddleware
from aiogram.types import Message
from aiogram.dispatcher.event.bases import SkipHandler


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

        self.user_message_times = defaultdict(lambda: deque())
        self.banned_users = {}  # user_id: ban_until_time

    async def __call__(self, handler, event: Message, data: dict):
        user_id = event.from_user.id
        now = time.monotonic()

        # проверка на бан
        ban_until = self.banned_users.get(user_id)
        if ban_until and now < ban_until:
            await event.answer("🚫 Ты временно забанен за спам. Подожди немного.")
            raise SkipHandler()

        times = self.user_message_times[user_id]
        times.append(now)

        # очищаем старые события
        while times and now - times[0] > self.limit_seconds:
            times.popleft()

        if len(times) > self.limit_count:
            # баним пользователя
            self.banned_users[user_id] = now + self.ban_seconds
            await event.answer("🚫 Слишком много сообщений! Ты временно забанен.")
            raise SkipHandler()

        # пропускаем к следующему хендлеру
        return await handler(event, data)
