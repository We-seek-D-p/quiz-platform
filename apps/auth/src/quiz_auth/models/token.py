from apps.auth.src.quiz_auth.models.users import UserPublic
from sqlmodel import SQLModel


class AccessToken(SQLModel):
    access_token: str
    token_type: str = "Bearer"  # noqa: S105
    expires_in: int


class LoginResponse(AccessToken):
    user: UserPublic
