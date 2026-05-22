import os
from openai import OpenAI

_base = os.getenv("OPENAI_BASE_URL", "https://api.deepseek.com/v1")
_key = os.getenv("OPENAI_API_KEY", "sk-placeholder")
_model = os.getenv("OPENAI_MODEL", "deepseek-chat")

_client = OpenAI(base_url=_base, api_key=_key)


def chat(system: str, user: str, temperature: float = 0.3) -> str:
    resp = _client.chat.completions.create(
        model=_model,
        messages=[
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        ],
        temperature=temperature,
    )
    return resp.choices[0].message.content or ""
