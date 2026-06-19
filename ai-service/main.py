import asyncio

from config.settings import settings
from grpcserver.grpc import Server


async def main() -> None:
    await Server(settings).run()


if __name__ == "__main__":
    asyncio.run(main())
