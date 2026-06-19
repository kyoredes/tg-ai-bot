import json
import re


_SSE_DATA_RE = re.compile(r"^data:\s*", re.MULTILINE)


def is_invalid_llm_response(text: str | None) -> bool:
    if not text or not text.strip():
        return True

    stripped = text.strip()

    if stripped.startswith("data:") or "[DONE]" in stripped:
        return True

    if _SSE_DATA_RE.search(stripped):
        return True

    lower = stripped.lower()

    if lower.startswith("error:"):
        return True

    if stripped.startswith("<!DOCTYPE") or stripped.startswith("<html"):
        return True

    if '"type":"error"' in stripped or '"errortext"' in lower:
        return True

    try:
        if stripped.startswith("{"):
            payload = json.loads(stripped)
            if isinstance(payload, dict) and payload.get("type") == "error":
                return True
    except json.JSONDecodeError:
        pass

    if stripped.startswith("<audio") or stripped.startswith("<video") or stripped.startswith("<img"):
        return True

    if "/media/" in stripped or "generated_media" in lower or "generated_images" in lower:
        return True

    if any(ext in lower for ext in (".mp3", ".wav", ".ogg", ".webm", ".mp4", ".png", ".jpg", ".jpeg", ".webp")):
        return True

    technical_markers = (
        "authentication error",
        "ошибка аутентификации",
        "no api key",
        "api key passed",
        "api ключ",
        "traceback",
        "exception",
        "status code",
        "please [log in]",
        "model not found",
        "не могу обработать ваш запрос",
    )
    return any(marker in lower for marker in technical_markers)
