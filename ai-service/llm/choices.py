from enum import Enum


class ProviderChoiceGPT(str, Enum):
    OPENAI = "openai"
    G4F = "g4f"
    ANY = "any"
