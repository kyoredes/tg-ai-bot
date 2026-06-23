import asyncio
import json
import logging
from dataclasses import dataclass

from aiokafka import AIOKafkaConsumer, AIOKafkaProducer

from config.kafka import TOPIC_PROFILE_ANALYZE_REQUESTS, TOPIC_PROFILE_ANALYZE_RESULTS, kafka_settings
from config.prompts import LLM_FALLBACK_ERROR_MESSAGE
from llm.profile_analyzer import ProfileData
from llm.profile_service import run_profile_analysis

logger = logging.getLogger(__name__)


@dataclass(slots=True)
class ProfileAnalyzeJob:
    job_id: str
    telegram_id: str
    chat_id: int
    progress_message_id: int
    first_name: str
    last_name: str = ""
    username: str = ""
    bio: str = ""
    is_premium: bool = False
    language_code: str = ""
    photo_base64: str = ""

    @classmethod
    def from_payload(cls, payload: dict) -> "ProfileAnalyzeJob":
        return cls(
            job_id=str(payload.get("jobId", "")),
            telegram_id=str(payload.get("telegramID", "")),
            chat_id=int(payload.get("chatID", 0)),
            progress_message_id=int(payload.get("progressMessageID", 0)),
            first_name=str(payload.get("firstName", "")),
            last_name=str(payload.get("lastName", "")),
            username=str(payload.get("username", "")),
            bio=str(payload.get("bio", "")),
            is_premium=bool(payload.get("isPremium", False)),
            language_code=str(payload.get("languageCode", "")),
            photo_base64=str(payload.get("photoBase64", "")),
        )

    def to_profile_data(self) -> ProfileData:
        return ProfileData(
            first_name=self.first_name,
            last_name=self.last_name,
            username=self.username,
            bio=self.bio,
            is_premium=self.is_premium,
            language_code=self.language_code,
            photo_base64=self.photo_base64,
        )

    def is_valid(self) -> bool:
        return bool(self.job_id and self.telegram_id and self.first_name and self.chat_id)


class ProfileAnalyzeWorker:
    def __init__(self) -> None:
        self._brokers = kafka_settings.broker_list()
        self._stop = asyncio.Event()

    async def run(self) -> None:
        if not self._brokers:
            logger.error("Kafka brokers are not configured, profile worker stopped")
            return

        while not self._stop.is_set():
            consumer = AIOKafkaConsumer(
                TOPIC_PROFILE_ANALYZE_REQUESTS,
                bootstrap_servers=self._brokers,
                group_id=kafka_settings.GROUP_ID,
                enable_auto_commit=False,
                auto_offset_reset="earliest",
            )
            producer = AIOKafkaProducer(
                bootstrap_servers=self._brokers,
                max_request_size=4 * 1024 * 1024,
            )
            await consumer.start()
            await producer.start()
            logger.info(
                "Profile analyze worker started (brokers=%s, topic=%s)",
                self._brokers,
                TOPIC_PROFILE_ANALYZE_REQUESTS,
            )

            try:
                async for message in consumer:
                    if self._stop.is_set():
                        break
                    await self._handle_message(consumer, producer, message)
            except asyncio.CancelledError:
                raise
            except Exception:
                logger.exception("Profile analyze worker crashed, restarting in 5s")
                await asyncio.sleep(5)
            finally:
                await consumer.stop()
                await producer.stop()

    def stop(self) -> None:
        self._stop.set()

    async def _publish_result(
        self,
        producer: AIOKafkaProducer,
        job: ProfileAnalyzeJob,
        response: str,
    ) -> bool:
        result = {
            "jobId": job.job_id,
            "telegramID": job.telegram_id,
            "chatID": job.chat_id,
            "progressMessageID": job.progress_message_id,
            "status": "ok",
            "response": response,
        }
        await producer.send_and_wait(
            TOPIC_PROFILE_ANALYZE_RESULTS,
            json.dumps(result, ensure_ascii=False).encode("utf-8"),
            key=job.telegram_id.encode("utf-8"),
        )
        return True

    async def _handle_message(
        self,
        consumer: AIOKafkaConsumer,
        producer: AIOKafkaProducer,
        message,
    ) -> None:
        raw = message.value
        job: ProfileAnalyzeJob | None = None
        response = LLM_FALLBACK_ERROR_MESSAGE
        published = False

        try:
            payload = json.loads(raw.decode("utf-8"))
            job = ProfileAnalyzeJob.from_payload(payload)
        except Exception:
            logger.exception("Invalid profile analyze job payload")
            await consumer.commit()
            return

        if not job.is_valid():
            logger.error(
                "Profile analyze job missing required fields: %s",
                raw.decode("utf-8", errors="replace")[:500],
            )
            if job.chat_id:
                try:
                    published = await self._publish_result(
                        producer,
                        job,
                        LLM_FALLBACK_ERROR_MESSAGE,
                    )
                except Exception:
                    logger.exception(
                        "Failed to publish validation error for job %s",
                        job.job_id,
                    )
            if published or not job.chat_id:
                await consumer.commit()
            return

        try:
            logger.info(
                "Processing profile analyze job %s for telegram_id=%s",
                job.job_id,
                job.telegram_id,
            )
            response = await run_profile_analysis(job.telegram_id, job.to_profile_data())
            published = await self._publish_result(producer, job, response)
            logger.info("Profile analyze job %s completed", job.job_id)
        except Exception:
            logger.exception("Profile analyze job %s failed", job.job_id)
            try:
                published = await self._publish_result(
                    producer,
                    job,
                    LLM_FALLBACK_ERROR_MESSAGE,
                )
            except Exception:
                logger.exception("Failed to publish profile error for job %s", job.job_id)
        finally:
            if published:
                await consumer.commit()
