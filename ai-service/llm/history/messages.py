from typing import Literal

from langchain_core.messages import AIMessage, BaseMessage, HumanMessage, SystemMessage
from pydantic import BaseModel


class ChatMessage(BaseModel):
    role: Literal["user", "assistant"]
    content: str


def to_openai_messages(
    system_prompt: str,
    messages: list[ChatMessage],
) -> list[dict[str, str]]:
    return [
        {"role": "system", "content": system_prompt},
        *[message.model_dump() for message in messages],
    ]


def to_langchain_messages(
    system_prompt: str,
    messages: list[ChatMessage],
) -> list[BaseMessage]:
    result: list[BaseMessage] = [SystemMessage(content=system_prompt)]
    for message in messages:
        if message.role == "user":
            result.append(HumanMessage(content=message.content))
        else:
            result.append(AIMessage(content=message.content))
    return result
