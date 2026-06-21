from __future__ import annotations

import grpc
from grpc import aio

from ratelimit.limiter import SlidingWindowLimiter


class RateLimitInterceptor(aio.ServerInterceptor):
    def __init__(self, limiter: SlidingWindowLimiter):
        self._limiter = limiter

    async def intercept_service(self, continuation, handler_call_details):
        key = handler_call_details.method
        if not self._limiter.allow(key):
            async def reject(request, context):
                await context.abort(
                    grpc.StatusCode.RESOURCE_EXHAUSTED,
                    "rate limit exceeded",
                )

            return grpc.unary_unary_rpc_method_handler(reject)

        return await continuation(handler_call_details)
