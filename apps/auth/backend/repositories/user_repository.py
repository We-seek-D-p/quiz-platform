from typing import Any, cast
from uuid import UUID

from fastapi import HTTPException
from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from backend.models.users import User, UserCreate, utcnow


USER_TABLE: Any = cast(Any, User).__table__
_DELETED_AT_COLUMN = USER_TABLE.c.deleted_at


class UserRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def get_by_id(self, user_id: UUID) -> User | None:
        query = select(User).where(User.id == user_id, _DELETED_AT_COLUMN.is_(None))
        result = await self.db.exec(query)
        return result.first()

    async def get_by_email(self, email: str) -> User | None:
        query = select(User).where(User.email == email, _DELETED_AT_COLUMN.is_(None))
        result = await self.db.exec(query)
        return result.first()

    async def get_existing_by_email(self, email: str) -> User | None:
        query = select(User).where(User.email == email)
        result = await self.db.exec(query)
        return result.first()

    async def get_by_nickname(self, nickname: str) -> User | None:
        query = select(User).where(User.nickname == nickname, _DELETED_AT_COLUMN.is_(None))
        result = await self.db.exec(query)
        return result.first()

    async def get_existing_by_nickname(self, nickname: str) -> User | None:
        query = select(User).where(User.nickname == nickname)
        result = await self.db.exec(query)
        return result.first()

    async def create_user(self, user_data: UserCreate, password_hash: str) -> User:
        user_data_dict = user_data.model_dump(exclude={"password"})
        user_data_dict["email"] = str(user_data_dict["email"]).lower()
        user = User(**user_data_dict, password_hash=password_hash)
        self.db.add(user)
        await self.db.commit()
        await self.db.refresh(user)
        return user

    async def update_user(self, user_id: UUID, user_data: dict) -> User:
        user = await self.get_by_id(user_id)
        if not user:
            raise HTTPException(status_code=404, detail="User not found")
        for key, value in user_data.items():
            setattr(user, key, value)
        self.db.add(user)
        await self.db.commit()
        await self.db.refresh(user)
        return user

    async def delete_user(self, user_id: UUID) -> bool:
        user = await self.get_by_id(user_id)
        if not user:
            raise HTTPException(status_code=404, detail="User not found")
        user.deleted_at = utcnow()
        self.db.add(user)
        await self.db.commit()
        return True

    async def increment_token_version(self, user_id: UUID) -> None:
        user = await self.get_by_id(user_id)
        if not user:
            return
        user.token_version += 1
        self.db.add(user)
        await self.db.commit()

    async def update_last_login(self, user_id: UUID) -> None:
        user = await self.get_by_id(user_id)
        if not user:
            return
        user.last_login_at = utcnow()
        self.db.add(user)
        await self.db.commit()
