import os
from models import AISettings

_settings: AISettings = AISettings(
    base_url=os.getenv("OPENAI_BASE_URL", "https://api.deepseek.com/v1"),
    api_key=os.getenv("OPENAI_API_KEY", "sk-placeholder"),
    model=os.getenv("OPENAI_MODEL", "deepseek-chat"),
)


def get_settings() -> AISettings:
    s = _settings
    return AISettings(base_url=s.base_url, api_key="••••" + s.api_key[-4:] if len(s.api_key) > 4 else "••••", model=s.model)


def update_settings(settings: AISettings) -> None:
    global _settings
    if settings.base_url:
        _settings.base_url = settings.base_url
    if settings.api_key and settings.api_key != "":
        _settings.api_key = settings.api_key
    if settings.model:
        _settings.model = settings.model


def reload_llm_client() -> None:
    import services.llm as llm
    from openai import OpenAI
    llm._client = OpenAI(base_url=_settings.base_url, api_key=_settings.api_key)
    llm._model = _settings.model
