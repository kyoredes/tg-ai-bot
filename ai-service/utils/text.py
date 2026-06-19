import re


async def truncate_text_by_words(text: str, max_words: int = 32) -> str:
    words = text.split()
    snippet = " ".join(words[:max_words])
    match = re.search(r"([.!?])[^.!?]*$", snippet)
    if match:
        return snippet[: match.end(1)]
    return snippet
