from pydantic import BaseModel


class UserModel(BaseModel):
    status: str


class ClientModel(BaseModel):
    tg_id: str
    user_id: str | None = None
    email: str | None = None
