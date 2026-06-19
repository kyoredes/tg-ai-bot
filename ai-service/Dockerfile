FROM python:3.13-slim

WORKDIR /app

RUN pip install --no-cache-dir uv grpcio-tools

COPY proto/ /proto
COPY ai-service/ .

RUN bash scripts/gen_proto.sh

RUN uv sync --no-dev

CMD ["uv", "run", "python", "main.py"]
