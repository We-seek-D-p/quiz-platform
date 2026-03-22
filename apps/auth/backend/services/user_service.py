from uuid import UUID

from fastapi import HTTPException
from sqlmodel.ext.asyncio.session import AsyncSession

from backend.models.users import User
from backend.repositories.user_repository import UserRepository
from backend.utils.security import hash_password


class UserService:
    def __init__(self, db: AsyncSession):
        self.user_repo = UserRepository(db)

    async def get_user_by_id(self, user_id: UUID, current_user: User) -> User:
        if user_id != current_user.id:
            raise HTTPException(status_code=403, detail="Not enough permissions")
        user = await self.user_repo.get_by_id(user_id)
        if not user:
            raise HTTPException(status_code=404, detail="User not found")
        return user

    async def update_user(self, update_data: dict, user_id: UUID, current_user: User) -> User:
        if user_id != current_user.id:
            raise HTTPException(status_code=403, detail="Not enough permissions")

        if "email" in update_data and update_data["email"]:
            existing = await self.user_repo.get_existing_by_email(update_data["email"])
            if existing and existing.id != user_id:
                raise HTTPException(400, "User with this email already exist")

        if "nickname" in update_data and update_data["nickname"]:
            existing = await self.user_repo.get_existing_by_nickname(update_data["nickname"])
            if existing and existing.id != user_id:
                raise HTTPException(400, "User with this nickname already exist")
        password_updated = False
        if "password" in update_data and update_data["password"]:
            password_updated = True
            update_data["password_hash"] = hash_password(update_data.pop("password"))

        user = await self.user_repo.update_user(update_data, user_id)
        if not user:
            raise HTTPException(404, "User not found")
        if password_updated:
            await self.user_repo.increment_token_version(user_id)

        return user

    async def delete_user(self, current_user: User, user_id: UUID) -> None:
        if user_id != current_user.id:
            raise HTTPException(status_code=403, detail="Not enough permissions")
        deleted = await self.user_repo.delete_user(user_id)
        if not deleted:
            raise HTTPException(404, "User not found")
