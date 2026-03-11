from fastapi import HTTPException

from sqlmodel.ext.asyncio.session import AsyncSession

from backend.models.users import UserCreate, User, UserLogin
from backend.repositories.user_repository import UserRepository
from backend.utils.security import hash_password, verify_password, create_tokens


class AuthService:
    def __init__(self, db: AsyncSession):
        self.user_repo = UserRepository(db)

    async def registry_user(self, user_data: UserCreate) -> User:
        if await self.user_repo.get_by_username(user_data.username):
            raise HTTPException(status_code=404, detail="User with this username already exist")
        if await self.user_repo.get_by_email(str(user_data.email)):
            raise HTTPException(status_code=404, detail="User with this email already exist")
        hashed_password = hash_password(user_data.password)
        user = await self.user_repo.create_user(user_data, hashed_password)
        return user

    async def login_user(self, user_to_login: UserLogin) -> dict[str, str]:
        user = await self.user_repo.get_by_username(user_to_login.username)
        if not user or not verify_password(user_to_login.password, user.hashed_password):
            raise HTTPException(status_code=401, detail="Incorrect username or password")
        await self.user_repo.update_last_login(user.id)

        return create_tokens(user.id, user.token_version)
