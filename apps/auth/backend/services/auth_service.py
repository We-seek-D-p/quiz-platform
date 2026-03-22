from fastapi import HTTPException, status

from sqlmodel.ext.asyncio.session import AsyncSession

from backend.models.users import UserCreate, User, UserLogin
from backend.repositories.role_repository import RoleRepository
from backend.repositories.user_repository import UserRepository
from backend.utils.security import hash_password, verify_password, create_tokens, decode_token


class AuthService:
    def __init__(self, db: AsyncSession):
        self.user_repo = UserRepository(db)
        self.role_repo = RoleRepository(db)

    async def registry_user(self, user_data: UserCreate) -> User:
        await self.role_repo.ensure_host_role()

        email = str(user_data.email).lower()
        if await self.user_repo.get_existing_by_nickname(user_data.nickname):
            raise HTTPException(status_code=409, detail="User with this nickname already exist")
        if await self.user_repo.get_existing_by_email(email):
            raise HTTPException(status_code=409, detail="User with this email already exist")
        normalized_payload = user_data.model_copy(update={"email": email})
        password_hash = hash_password(user_data.password)
        user = await self.user_repo.create_user(normalized_payload, password_hash)
        return user

    async def login_user(self, user_to_login: UserLogin) -> dict[str, str]:
        email = str(user_to_login.email).lower()
        user = await self.user_repo.get_by_email(email)
        if not user or not verify_password(user_to_login.password, user.password_hash):
            raise HTTPException(status_code=401, detail="Incorrect email or password")
        await self.user_repo.update_last_login(user.id)

        return create_tokens(user.id, user.token_version)

    async def refresh_tokens(self, refresh_token: str) -> dict:
        user_id, _ = decode_token(refresh_token, expected_type="refresh")
        if not user_id:
            raise HTTPException(status_code=401, detail="Invalid refresh token")
        user = await self.user_repo.get_by_id(user_id)
        if not user:
            raise HTTPException(status_code=404, detail="User not found")
        return create_tokens(user.id, user.token_version)
