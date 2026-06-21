import asyncio
import logging
import signal

import grpc

from config.settings import Settings
from config.throttle import throttle_settings
from grpcserver.server import AIServer
from ratelimit.interceptor import RateLimitInterceptor
from ratelimit.limiter import SlidingWindowLimiter
from rpc.ai.v1 import ai_pb2_grpc


class Server:
    def __init__(self, settings: Settings):
        self._settings = settings
        interceptors = []
        if throttle_settings.ENABLED:
            interceptors.append(
                RateLimitInterceptor(
                    SlidingWindowLimiter(
                        throttle_settings.LIMIT,
                        throttle_settings.WINDOW_SECONDS,
                    ),
                ),
            )
        self._grpc_server = grpc.aio.server(interceptors=interceptors)

    async def run(self) -> None:
        logging.basicConfig(level=self._settings.LOG_LEVEL)
        ai_pb2_grpc.add_AIServiceServicer_to_server(AIServer(), self._grpc_server)

        listen_addr = f"{self._settings.GRPC_HOST}:{self._settings.GRPC_PORT}"
        self._grpc_server.add_insecure_port(listen_addr)
        await self._grpc_server.start()
        logging.info("gRPC server started on %s", listen_addr)

        stop_event = asyncio.Event()

        def _shutdown(*_: object) -> None:
            stop_event.set()

        loop = asyncio.get_running_loop()
        for sig in (signal.SIGINT, signal.SIGTERM):
            loop.add_signal_handler(sig, _shutdown)

        await stop_event.wait()
        await self._grpc_server.stop(grace=5)
        logging.info("gRPC server stopped")
