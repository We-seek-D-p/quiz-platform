from uuid import UUID

from fastapi import HTTPException
from sqlmodel.ext.asyncio.session import AsyncSession

from backend.models.users import User, UserUpdate
from backend.repositories.user_repository import UserRepository
from backend.utils.security import hash_password, verify_password


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

    async def update_profile(self, user_id: UUID, update_data: UserUpdate) -> User:
        update_dict = update_data.model_dump(exclude={"password", "role"}, exclude_unset=True)

        if not update_dict:
            raise HTTPException(400, "No profile data to update")

        if "email" in update_dict:
            email = update_dict["email"].lower()
            existing = await self.user_repo.get_existing_by_email(email)
            if existing and existing.id != user_id:
                raise HTTPException(409, "User with this email already exists")
            update_dict["email"] = email

        if "nickname" in update_dict:
            existing = await self.user_repo.get_existing_by_nickname(update_dict["nickname"])
            if existing and existing.id != user_id:
                raise HTTPException(409, "User with this nickname already exists")

        user = await self.user_repo.update_user(user_id, update_dict)
        if not user:
            raise HTTPException(404, "User not found")
        return user

    async def change_password(self, user_id: UUID, current_password: str, new_password: str) -> None:
        user = await self.user_repo.get_by_id(user_id)
        if not user or not verify_password(current_password, user.password_hash):
            raise HTTPException(status_code=401, detail="Incorrect current password")

        new_hash = hash_password(new_password)
        await self.user_repo.update_user(user_id, {"password_hash": new_hash})

        # Выкидываем изо всех сессий
        await self.user_repo.increment_token_version(user_id)

    async def delete_user(self, current_user: User, user_id: UUID) -> None:
        if user_id != current_user.id:
            raise HTTPException(status_code=403, detail="Not enough permissions")
        deleted = await self.user_repo.delete_user(user_id)
        if not deleted:
            raise HTTPException(404, "User not found")
