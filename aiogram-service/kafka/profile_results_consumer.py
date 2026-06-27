import asyncio
import json
import logging

from aiokafka import AIOKafkaConsumer
from aiogram import Bot
from aiogram.exceptions import TelegramAPIError

from config.core import settings
from core.keyboards import get_back_keyboard
from users.answers import PROFILE_ROAST_ERROR_ANSWER
from utils.response import is_user_safe_response, split_telegram_message

logger = logging.getLogger(__name__)

TOPIC_PROFILE_ANALYZE_RESULTS = "profile.analyze.results"


async def consume_profile_results(bot: Bot) -> None:
    while True:
        try:
            await _run_consumer(bot)
        except asyncio.CancelledError:
            raise
        except Exception:
            logger.exception("Profile result consumer crashed, restarting in 5s")
            await asyncio.sleep(5)


async def _run_consumer(bot: Bot) -> None:
    brokers = settings.kafka_broker_list()
    if not brokers:
        logger.error("Kafka brokers are not configured, profile result consumer stopped")
        return

    consumer = AIOKafkaConsumer(
        TOPIC_PROFILE_ANALYZE_RESULTS,
        bootstrap_servers=brokers,
        group_id=settings.KAFKA_GROUP_ID,
        enable_auto_commit=True,
        auto_offset_reset="latest",
    )
    await consumer.start()
    logger.info(
        "Profile result consumer started (brokers=%s, topic=%s)",
        brokers,
        TOPIC_PROFILE_ANALYZE_RESULTS,
    )

    try:
        async for message in consumer:
            await _handle_result(bot, message.value)
    finally:
        await consumer.stop()


async def _handle_result(bot: Bot, raw: bytes) -> None:
    try:
        payload = json.loads(raw.decode("utf-8"))
        chat_id = int(payload.get("chatID", 0))
        response_text = str(payload.get("response", "")).strip()
        progress_message_id = int(payload.get("progressMessageID", 0))
        job_id = str(payload.get("jobId", ""))
    except Exception as exc:
        logger.error("Invalid profile result payload: %s", exc)
        return

    if not chat_id:
        logger.error("Profile result payload missing chatID (job_id=%s)", job_id)
        return

    if progress_message_id:
        try:
            await bot.delete_message(chat_id=chat_id, message_id=progress_message_id)
        except TelegramAPIError:
            pass

    if not response_text or not is_user_safe_response(response_text):
        logger.error("Unsafe or empty profile result for job_id=%s", job_id)
        await _send_text(bot, chat_id, PROFILE_ROAST_ERROR_ANSWER)
        return

    await _send_text(bot, chat_id, response_text)


async def _send_text(bot: Bot, chat_id: int, text: str) -> None:
    parts = split_telegram_message(text)
    for index, part in enumerate(parts):
        try:
            await bot.send_message(
                chat_id=chat_id,
                text=part,
                reply_markup=get_back_keyboard() if index == len(parts) - 1 else None,
            )
        except TelegramAPIError:
            logger.exception("Failed to send profile result to chat_id=%s", chat_id)
            if index == 0:
                try:
                    await bot.send_message(
                        chat_id=chat_id,
                        text=PROFILE_ROAST_ERROR_ANSWER,
                        reply_markup=get_back_keyboard(),
                    )
                except TelegramAPIError:
                    logger.exception(
                        "Failed to send profile error fallback to chat_id=%s",
                        chat_id,
                    )
            return
