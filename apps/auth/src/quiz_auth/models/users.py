import re
from datetime import UTC, datetime
from uuid import UUID, uuid7

from pydantic import EmailStr, field_validator
from sqlalchemy import Column, DateTime, ForeignKey, Index, String, func
from sqlmodel import Field, SQLModel
from zxcvbn import zxcvbn

AUTH_SCHEMA = "auth"
RESERVED_USERNAMES = {
    "admin",
    "support",
    "root",
    "settings",
    "api",
    "auth",
    "login",
    "logout",
    "signup",
    "register",
    "password",
    "management",
    "owner",
    "superuser",
}
USERNAME_RE = re.compile(r"^[a-zA-Z0-9_-]+$")


def utcnow() -> datetime:
    return datetime.now(UTC)


class Role(SQLModel, table=True):
    __tablename__ = "roles"
    __table_args__ = {"schema": AUTH_SCHEMA}

    slug: str = Field(primary_key=True, max_length=50)
    name: str = Field(max_length=100)
    priority: int = Field(default=0)


class UserBase(SQLModel):
    nickname: str = Field(unique=True, index=True)
    email: EmailStr = Field(unique=True, index=True)


class UserPublic(UserBase):
    id: UUID
    role: str


class User(UserBase, table=True):
    __tablename__ = "users"
    __table_args__ = (
        Index("ix_auth_users_role", "role"),
        Index("ix_auth_users_deleted_at", "deleted_at"),
        {"schema": AUTH_SCHEMA},
    )

    id: UUID = Field(default_factory=uuid7, primary_key=True)
    password_hash: str

    role: str = Field(
        default="host",
        sa_column=Column(
            String(50),
            ForeignKey(f"{AUTH_SCHEMA}.roles.slug", onupdate="CASCADE"),
            nullable=False,
            server_default="host",
        ),
    )

    token_version: int = Field(default=0)

    created_at: datetime = Field(
        default_factory=utcnow,
        sa_column=Column(
            DateTime(timezone=True),
            nullable=False,
            server_default=func.now(),
        ),
    )
    updated_at: datetime = Field(
        default_factory=utcnow,
        sa_column=Column(
            DateTime(timezone=True),
            nullable=False,
            server_default=func.now(),
            onupdate=func.now(),
        ),
    )
    last_login_at: datetime | None = Field(
        default=None,
        sa_column=Column(DateTime(timezone=True), nullable=True),
    )
    deleted_at: datetime | None = Field(
        default=None,
        sa_column=Column(DateTime(timezone=True), nullable=True),
    )


class RefreshToken(SQLModel, table=True):
    __tablename__ = "refresh_tokens"
    __table_args__ = (
        Index("ix_auth_refresh_tokens_user_id", "user_id"),
        Index("ix_auth_refresh_tokens_expires_at", "expires_at"),
        {"schema": AUTH_SCHEMA},
    )

    id: UUID = Field(default_factory=uuid7, primary_key=True)

    user_id: UUID = Field(
        sa_column=Column(
            ForeignKey(f"{AUTH_SCHEMA}.users.id", onupdate="CASCADE"),
            nullable=False,
        )
    )

    token_hash: str = Field(max_length=64, unique=True)

    issued_at: datetime = Field(
        default_factory=utcnow,
        sa_column=Column(
            DateTime(timezone=True),
            nullable=False,
            server_default=func.now(),
        ),
    )
    expires_at: datetime = Field(
        sa_column=Column(
            DateTime(timezone=True),
            nullable=False,
        )
    )
    revoked_at: datetime | None = Field(
        default=None,
        sa_column=Column(DateTime(timezone=True), nullable=True),
    )


class UserCreate(UserBase):
    password: str

    @field_validator("nickname")
    @classmethod
    def validate_nickname(cls, v: str):
        v = v.lower()
        if v in RESERVED_USERNAMES:
            raise ValueError("This nickname is reserved by the system")
        if len(v) < 3:
            raise ValueError("The nickname is too short")
        if not re.match(USERNAME_RE, v):
            raise ValueError("Only Latin characters, numbers, '-' and '_' are allowed")
        return v

    @field_validator("password")
    @classmethod
    def validate_password_strength(cls, v: str, info):
        user_inputs = []
        if "nickname" in info.data:
            user_inputs.append(info.data["nickname"])
        if "email" in info.data:
            user_inputs.append(info.data["email"])

        results = zxcvbn(v, user_inputs=user_inputs)

        if results["score"] < 3:
            feedback = results["feedback"]["warning"]
            suggestions = ". ".join(results["feedback"]["suggestions"])
            error_msg = f"The password is too weak. {feedback}. {suggestions}"
            raise ValueError(error_msg)

        return v


class UserLogin(SQLModel):
    email: EmailStr
    password: str


class UserUpdate(SQLModel):
    nickname: str | None = None
    email: EmailStr | None = None
    password: str | None = None
    role: str | None = None
