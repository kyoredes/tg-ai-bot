G4F_FALLBACK_ERROR_MESSAGE = (
    "Сейчас не удалось получить ответ от нейросети. "
    "Пожалуйста, попробуйте ещё раз через несколько минут. "
    "Если ошибка повторится — напишите в поддержку."
)


class LLMUserFacingError(Exception):
    """Ошибка с текстом, безопасным для показа пользователю."""

    def __init__(self, message: str):
        self.message = message
        super().__init__(message)
