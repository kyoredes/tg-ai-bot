import asyncio

from config.settings import settings
from grpcserver.grpc import Server
from kafka.profile_worker import ProfileAnalyzeWorker


async def main() -> None:
    worker = ProfileAnalyzeWorker()
    await asyncio.gather(
        Server(settings).run(),
        worker.run(),
    )


if __name__ == "__main__":
    asyncio.run(main())
