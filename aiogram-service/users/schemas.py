from pydantic import BaseModel


class UserModel(BaseModel):
    status: str


class ClientModel(BaseModel):
    tg_id: str
    user_id: str | None = None
    email: str | None = None


class SubscriptionModel(BaseModel):
    subscription_id: str
    user_id: str
    starts_at: int
    expires_at: int


class ChatModel(BaseModel):
    tg_id: str
    response: str
