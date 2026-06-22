def is_user_safe_response(text: str) -> bool:
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


def split_telegram_message(text: str, *, limit: int = 4000) -> list[str]:
    if len(text) <= limit:
        return [text]
    return [text[i : i + limit] for i in range(0, len(text), limit)]
