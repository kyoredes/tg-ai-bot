from dataclasses import dataclass
from typing import Any

from g4f.Provider import OperaAria, RetryProvider, Yqcloud


@dataclass(frozen=True, slots=True)
class G4FModelConfig:
    name: str
    label: str = ""
    provider: Any = None

    def __str__(self) -> str:
        return self.label or self.name


G4F_FALLBACK_MAX_ATTEMPTS = 4

# По доке: model + provider (классы), без ignore_stream и без строковых провайдеров.
# RetryProvider — рекомендованный способ fallback внутри g4f.
G4F_FALLBACK_MODELS: tuple[G4FModelConfig, ...] = (
    G4FModelConfig(
        "",
        "Retry Yqcloud + OperaAria",
        provider=RetryProvider([Yqcloud, OperaAria], shuffle=False),
    ),
    G4FModelConfig("gpt-4", "Yqcloud", provider=Yqcloud),
    G4FModelConfig("", "Opera Aria", provider=OperaAria),
    G4FModelConfig("gpt-4", "Yqcloud retry", provider=Yqcloud),
)

G4F_MODELS: tuple[G4FModelConfig, ...] = G4F_FALLBACK_MODELS
