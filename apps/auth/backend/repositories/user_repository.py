from datetime import datetime

from fastapi import HTTPException
from sqlmodel.ext.asyncio.session import AsyncSession
from sqlmodel import select, update

from backend.models.users import User, UserCreate, UserUpdate


class UserRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def get_by_id(self, user_id: int) -> User | None:
        query = select(User).where(User.id == user_id, User.is_active == True)
        result = await self.db.exec(query)
        return result.first()

    async def get_by_email(self, email: str) -> User | None:
        result = await self.db.exec(select(User).where(User.email == email, User.is_active == True))
        return result.first()

    async def get_by_username(self, username: str) -> User | None:
        result = await self.db.exec(select(User).where(User.username == username, User.is_active == True))
        return result.first()

    async def create_user(self, user_data: UserCreate, hashed_password: str) -> User:
        user_data_dict = user_data.model_dump(exclude={"password"})
        user = User(**user_data_dict, hashed_password=hashed_password)
        self.db.add(user)
        await self.db.commit()
        await self.db.refresh(user)
        return user

    async def update_user(self, user_data: dict, user_id: int) -> User:
        user = await self.get_by_id(user_id)
        if not user:
            raise HTTPException(status_code=404, detail="User not found")
        for key, value in user_data.items():
            setattr(user, key, value)
        self.db.add(user)
        await self.db.commit()
        await self.db.refresh(user)
        return user

    async def delete_user(self, user_id: int) -> bool:
        user = await self.get_by_id(user_id)
        if not user:
            raise HTTPException(status_code=404, detail="User not found")
        user.is_active = False
        self.db.add(user)
        await self.db.commit()
        return True

    async def increment_token_version(self, user_id: int) -> None:
        stmt = update(User).where(User.id == user_id, User.is_active == True).values(
            token_version=User.token_version + 1)
        await self.db.exec(stmt)
        await self.db.commit()

    async def update_last_login(self, user_id: int) -> None:
        stmt = update(User).where(User.id == user_id, User.is_active == True).values(last_login_at=datetime.now())
        await self.db.exec(stmt)
        await self.db.commit()
