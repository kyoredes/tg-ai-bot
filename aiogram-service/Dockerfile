FROM python:3.13-slim

WORKDIR /app

RUN pip install --no-cache-dir \
    "aiogram>=3.27.0" \
    "httpx>=0.28.1" \
    "pydantic-settings>=2.13.1" \
    "aiokafka>=0.12.0"

COPY aiogram-service/ .

CMD ["python", "main.py"]
