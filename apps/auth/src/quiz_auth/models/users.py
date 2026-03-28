from datetime import UTC, datetime
from uuid import UUID, uuid7

from pydantic import EmailStr
from sqlalchemy import Column, DateTime, ForeignKey, Index, String, func
from sqlmodel import Field, SQLModel

AUTH_SCHEMA = "auth"


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


class UserLogin(SQLModel):
    email: EmailStr
    password: str


class UserUpdate(SQLModel):
    nickname: str | None = None
    email: EmailStr | None = None
    password: str | None = None
    role: str | None = None
