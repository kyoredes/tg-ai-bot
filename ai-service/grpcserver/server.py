import logging

import grpc

from config.throttle import throttle_settings
from grpcserver.admin import AdminService
from llm.errors import LLMUserFacingError
from llm.manager import LLMManager
from ratelimit.limiter import SlidingWindowLimiter
from rpc.ai.v1 import ai_pb2, ai_pb2_grpc
from utils.response import is_invalid_llm_response

logger = logging.getLogger(__name__)

_admin_service = AdminService()
_chat_limiter = SlidingWindowLimiter(
    throttle_settings.CHAT_LIMIT,
    throttle_settings.CHAT_WINDOW_SECONDS,
)


class AIServer(ai_pb2_grpc.AIServiceServicer):
    async def Chat(self, request, context):
        telegram_id = request.telegram_id.strip()
        prompt = request.prompt.strip()

        if not telegram_id or not prompt:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("telegram_id and prompt are required")
            return ai_pb2.ChatResponse()

        if throttle_settings.ENABLED and not _chat_limiter.allow(f"chat:{telegram_id}"):
            await context.abort(
                grpc.StatusCode.RESOURCE_EXHAUSTED,
                "rate limit exceeded",
            )

        try:
            manager = LLMManager(session_id=telegram_id)
            response = await manager.make_request(prompt)
        except LLMUserFacingError as exc:
            logger.warning(
                "User-facing LLM error for telegram_id=%s: %s",
                telegram_id,
                exc.message,
            )
            return ai_pb2.ChatResponse(
                telegram_id=telegram_id,
                response=exc.message,
            )
        except Exception as exc:
            logger.exception("Chat failed for telegram_id=%s", telegram_id)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(exc))
            return ai_pb2.ChatResponse()

        if is_invalid_llm_response(response):
            logger.error("LLM returned invalid response for telegram_id=%s", telegram_id)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details("invalid llm response")
            return ai_pb2.ChatResponse()

        return ai_pb2.ChatResponse(
            telegram_id=telegram_id,
            response=response,
        )

    async def GetChatHistory(self, request, context):
        telegram_id = request.telegram_id.strip()
        if not telegram_id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("telegram_id is required")
            return ai_pb2.GetChatHistoryResponse()

        messages = await _admin_service.get_chat_history(telegram_id)
        return ai_pb2.GetChatHistoryResponse(
            telegram_id=telegram_id,
            messages=[
                ai_pb2.ChatMessage(role=m["role"], content=m["content"])
                for m in messages
            ],
        )

    async def ClearChatHistory(self, request, context):
        telegram_id = request.telegram_id.strip()
        if not telegram_id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("telegram_id is required")
            return ai_pb2.ClearChatHistoryResponse()

        await _admin_service.clear_chat_history(telegram_id)
        return ai_pb2.ClearChatHistoryResponse()

    async def ListChatSessions(self, request, context):
        sessions, total = await _admin_service.list_chat_sessions(
            request.page or 1,
            request.limit or 20,
        )
        return ai_pb2.ListChatSessionsResponse(
            sessions=[
                ai_pb2.ChatSessionItem(
                    telegram_id=s["telegram_id"],
                    message_count=s["message_count"],
                )
                for s in sessions
            ],
            total=total,
        )

    async def GetLLMConfig(self, request, context):
        config = _admin_service.get_llm_config()
        return ai_pb2.GetLLMConfigResponse(
            model=config["model"],
            temperature=config["temperature"],
            max_tokens=config["max_tokens"],
            debug=config["debug"],
            provider=config["provider"],
            g4f_models=config["g4f_models"],
            uses_litellm=config["uses_litellm"],
        )

    async def GetSystemPrompt(self, request, context):
        data = await _admin_service.get_system_prompt()
        return ai_pb2.GetSystemPromptResponse(
            prompt=data["prompt"],
            default_prompt=data["default_prompt"],
            is_custom=data["is_custom"],
        )

    async def UpdateSystemPrompt(self, request, context):
        data = await _admin_service.update_system_prompt(request.prompt)
        return ai_pb2.UpdateSystemPromptResponse(
            prompt=data["prompt"],
            default_prompt=data["default_prompt"],
            is_custom=data["is_custom"],
        )

    async def Health(self, request, context):
        data = await _admin_service.check_health()
        return ai_pb2.HealthResponse(
            ok=data["ok"],
            db_ok=data["db_ok"],
            redis_ok=data["redis_ok"],
        )
