from enum import Enum


class ProviderChoiceGPT(str, Enum):
    LITELLM = "litellm"
    G4F = "g4f"
    ANY = "any"
