from datetime import datetime

from pydantic import EmailStr
from sqlalchemy import Column, func, DateTime
from sqlmodel import SQLModel, Field


class UserBase(SQLModel):
    username: str = Field(index=True, unique=True)
    age: int = Field(index=True)
    email: EmailStr = Field(index=True, unique=True)


class UserPublic(UserBase):
    id: int


class User(UserBase, table=True):
    id: int | None = Field(primary_key=True, default=None)
    role: str = Field(default="player")
    hashed_password: str
    is_active: bool = Field(default=True)
    token_version: int = Field(default=1)
    created_at: datetime | None = Field(sa_column=Column(DateTime, server_default=func.now()))
    updated_at: datetime | None = Field(sa_column=Column(DateTime, server_default=func.now(), onupdate=func.now()))
    last_login_at: datetime | None = Field(default=None)



class UserCreate(UserBase):
    password: str

class UserLogin(SQLModel):
    username: str
    password: str


class UserUpdate(SQLModel):
    username: str | None = None
    age: int | None = None
    email: EmailStr | None = None
    password: str | None = None
    is_active: bool | None = None
