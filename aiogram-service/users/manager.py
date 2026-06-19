import httpx
from config.core import settings
import logging
from users.schemas import ClientModel, UserModel
from contextlib import asynccontextmanager


logger = logging.getLogger(__name__)


class UserManager:
    def __init__(self):
        self.backend_url = f"{settings.GATEWAY_HOST}:{settings.GATEWAY_PORT}"
        self.api_key = settings.COMMON_PUB_KEY

    @asynccontextmanager
    async def _get_client(self):
        async with httpx.AsyncClient() as client:
            yield client

    async def _get_headers(self) -> dict:
        return {
            "Authorization": self.api_key,
        }

    async def start(self, tg_id: str) -> UserModel | None:
        headers = await self._get_headers()
        logger.debug("Starting telegram with telegram id %s", tg_id)
        body = {
            "telegramID": tg_id,
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
                        "Unable to start telegram with client %s: status %s",
                        tg_id,
                        response.status_code,
                    )
                    return None
                result = response.json()
                logger.debug("Response %s", result)
                if not result:
                    return None
                if result.get("status") != "ok":
                    raise RuntimeError(f"Error gateway {result.get('status')}")
                return UserModel(status=result["status"])
            except Exception as e:
                logger.error("Error %s: %s", tg_id, e)
                return None

    async def get_profile(self, tg_id: str) -> ClientModel | None:
        headers = await self._get_headers()
        body = {
            "telegramID": tg_id,
        }
        async with self._get_client() as client:
            try:
                response = await client.post(
                    headers=headers,
                    url=f"http://{self.backend_url}/telegram/profile",
                    json=body,
                )
                if response.status_code != 200:
                    logger.error(
                        "Unable to get profile for client %s: status %s",
                        tg_id,
                        response.status_code,
                    )
                    return None
                result = response.json()
                if result.get("status") != "ok":
                    return None
                profile = result.get("profile") or {}
                return ClientModel(
                    tg_id=profile.get("telegramID", tg_id),
                    user_id=profile.get("userID"),
                    email=profile.get("email") or None,
                )
            except Exception as e:
                logger.error("Error getting profile %s: %s", tg_id, e)
                return None
