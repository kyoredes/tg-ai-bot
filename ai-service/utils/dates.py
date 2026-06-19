
from datetime import datetime, timezone


def check_dates(start_date: datetime | None, end_date: datetime | None) -> bool:
    """Проверяет, активна ли подписка.
    Возвращает True, если текущая дата между start_date и end_date (включительно).
    Если одна из дат не указана — возвращает False.
    """
    if not start_date and not end_date:
        return True

    now = datetime.now(timezone.utc)
    return start_date <= now <= end_date