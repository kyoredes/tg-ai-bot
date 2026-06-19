import re


def count_words(text: str) -> int:
    return len(text.split())


def truncate_to_word_limit(text: str, max_words: int) -> str:
    if max_words <= 0:
        return text

    words = text.split()
    if len(words) <= max_words:
        return text
    return " ".join(words[:max_words])


async def truncate_text_by_words(text: str, max_words: int = 32) -> str:
    words = text.split()
    snippet = " ".join(words[:max_words])
    match = re.search(r"([.!?])[^.!?]*$", snippet)
    if match:
        return snippet[: match.end(1)]
    return snippet
