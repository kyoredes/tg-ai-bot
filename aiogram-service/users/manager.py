import httpx
from config.core import settings
import logging
from users.schemas import UserModel
from contextlib import asynccontextmanager


logger = logging.getLogger(__name__)


class UserManager:
    def __init__(self):
        self.backend_url = f"{settings.GATEWAY_HOST}:{settings.GATEWAY_PORT}"
        self.api_key = settings.COMMON_PUB_KEY

    @asynccontextmanager
    async def _get_client(self):
        """Контекстный менеджер для правильного управления HTTP-клиентом"""
        async with httpx.AsyncClient() as client:
            yield client

    async def _get_headers(self) -> dict:
        headers = {
            "Authorization": f"{self.api_key}",
        }
        return headers

    async def start(self, tg_id: str) -> UserModel | None:
        headers = await self._get_headers()
        logger.debug(f"Starting telegram with telegram id {tg_id}")
        body = {
            'TelegramID': tg_id,
        }
        async with self._get_client() as client:
            try:
                response = await client.post(
                    headers=headers,
                    url=f"http://{self.backend_url}/telegram/start",
                    json=body,
                )
                if response.status_code != 200:
                    logger.error(
                        f"Unable to start telegram with client {tg_id}: status {response.status_code}"
                    )
                    return None
                result = response.json()
                logger.debug(f"Response {result}")
                if not result:
                    return None
                if result.get("status") != "ok":
                    raise Exception(f"Error gateway {result.get('status')}")
                client_model = UserModel(**result)
                return client_model
            except Exception as e:
                logger.error(f"Error {tg_id}: {e}")
                return None
