from config.settings import settings
from llm.choices import ProviderChoiceGPT
from llm.managers.g4f_manager import G4FManager
from llm.managers.litellm_manager import LiteLLMManager

providers_factory = {
    ProviderChoiceGPT.LITELLM: LiteLLMManager,
    ProviderChoiceGPT.G4F: G4FManager,
}

providers_factory_dev = {
    ProviderChoiceGPT.G4F: G4FManager,
}

PROVIDERS_FACTORY = providers_factory_dev if settings.DEBUG else providers_factory
