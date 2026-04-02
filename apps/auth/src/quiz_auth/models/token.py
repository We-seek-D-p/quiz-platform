from sqlmodel import SQLModel

from quiz_auth.models.users import UserPublic


class AccessToken(SQLModel):
    access_token: str
    token_type: str = "Bearer"  # noqa: S105
    expires_in: int


class LoginResponse(AccessToken):
    user: UserPublic
