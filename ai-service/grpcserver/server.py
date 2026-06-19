import logging

import grpc

from llm.errors import LLMUserFacingError
from llm.manager import LLMManager
from rpc.ai.v1 import ai_pb2, ai_pb2_grpc
from utils.response import is_invalid_llm_response

logger = logging.getLogger(__name__)


class AIServer(ai_pb2_grpc.AIServiceServicer):
    async def Chat(self, request, context):
        telegram_id = request.telegram_id.strip()
        prompt = request.prompt.strip()

        if not telegram_id or not prompt:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("telegram_id and prompt are required")
            return ai_pb2.ChatResponse()

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
