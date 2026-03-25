from sqlmodel import SQLModel

from backend.models.users import UserPublic


class AccessToken(SQLModel):
    access_token: str
    token_type: str = "Bearer"
    expires_in: int


class LoginResponse(AccessToken):
    user: UserPublic
