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


class ProfileSnapshot(BaseModel):
    first_name: str
    last_name: str = ""
    username: str = ""
    bio: str = ""
    is_premium: bool = False
    language_code: str = ""
    photo_base64: str | None = None
    has_photo: bool = False


class ProfileAnalyzeAcceptedModel(BaseModel):
    job_id: str

