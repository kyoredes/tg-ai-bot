from config.settings import settings
from llm.choices import ProviderChoiceGPT
from llm.managers.g4f_manager import G4FManager
from llm.managers.openai_manager import OpenAIManager

providers_factory = {
    ProviderChoiceGPT.OPENAI: OpenAIManager,
    ProviderChoiceGPT.G4F: G4FManager,
}

providers_factory_dev = {
    ProviderChoiceGPT.G4F: G4FManager,
}

PROVIDERS_FACTORY = providers_factory_dev if settings.DEBUG else providers_factory
