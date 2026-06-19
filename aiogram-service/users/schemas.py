from typing import List, Optional
from pydantic import BaseModel

class User(BaseModel):
    telegramID: str
    userID: str | None = None
    deviceID: str | None = None

class UserModel(BaseModel):
    status: str
    email: Optional[str] = None
