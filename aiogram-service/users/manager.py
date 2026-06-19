import httpx
from config.core import settings
import logging
from users.schemas import ChatModel, ClientModel, SubscriptionModel, UserModel
from contextlib import asynccontextmanager


logger = logging.getLogger(__name__)


def _is_user_safe_response(text: str) -> bool:
    stripped = (text or "").strip()
    if not stripped:
        return False
    lower = stripped.lower()
    if stripped.startswith("data:") or "[done]" in lower:
        return False
    if stripped.startswith("<!DOCTYPE") or stripped.startswith("<html"):
        return False
    if '"type":"error"' in stripped or "authentication error" in lower:
        return False
    if "api key" in lower and "error" in lower:
        return False
    return True


class UserManager:
    def __init__(self):
        self.backend_url = f"{settings.GATEWAY_HOST}:{settings.GATEWAY_PORT}"
        self.api_key = settings.COMMON_PUB_KEY

    @asynccontextmanager
    async def _get_client(self):
        async with httpx.AsyncClient(timeout=settings.HTTP_TIMEOUT) as client:
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

    async def get_subscription(self, tg_id: str) -> SubscriptionModel | None:
        headers = await self._get_headers()
        body = {"telegramID": tg_id}
        async with self._get_client() as client:
            try:
                response = await client.post(
                    headers=headers,
                    url=f"http://{self.backend_url}/telegram/subscription",
                    json=body,
                )
                if response.status_code != 200:
                    logger.error(
                        "Unable to get subscription for client %s: status %s",
                        tg_id,
                        response.status_code,
                    )
                    return None
                result = response.json()
                if result.get("status") != "ok":
                    return None
                sub = result.get("subscription") or {}
                return SubscriptionModel(
                    subscription_id=sub.get("subscriptionID", ""),
                    user_id=sub.get("userID", ""),
                    starts_at=sub.get("startsAt", 0),
                    expires_at=sub.get("expiresAt", 0),
                )
            except Exception as e:
                logger.error("Error getting subscription %s: %s", tg_id, e)
                return None

    async def chat(self, tg_id: str, prompt: str) -> ChatModel | None:
        headers = await self._get_headers()
        body = {"telegramID": tg_id, "prompt": prompt}
        async with self._get_client() as client:
            try:
                response = await client.post(
                    headers=headers,
                    url=f"http://{self.backend_url}/telegram/chat",
                    json=body,
                )
                if response.status_code != 200:
                    logger.error(
                        "Unable to chat for client %s: status %s",
                        tg_id,
                        response.status_code,
                    )
                    return None
                result = response.json()
                if result.get("status") != "ok":
                    return None
                chat = result.get("chat") or {}
                response = chat.get("response", "")
                if not _is_user_safe_response(response):
                    logger.error("Unsafe chat response for client %s", tg_id)
                    return None
                return ChatModel(
                    tg_id=chat.get("telegramID", tg_id),
                    response=response,
                )
            except Exception as e:
                logger.error("Error in chat %s: %s", tg_id, e)
                return None
