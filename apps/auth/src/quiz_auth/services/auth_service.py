from datetime import datetime, timezone
from uuid import UUID

from fastapi import HTTPException, status

from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_auth.models.users import UserCreate, User, UserLogin
from quiz_auth.repositories.refresh_token_repository import RefreshTokenRepository
from quiz_auth.repositories.role_repository import RoleRepository
from quiz_auth.repositories.user_repository import UserRepository
from quiz_auth.utils.security import (
    TokenPair,
    create_tokens,
    decode_token,
    hash_password,
    hash_refresh_token,
    verify_password,
)


class AuthService:
    def __init__(self, db: AsyncSession):
        self.user_repo = UserRepository(db)
        self.role_repo = RoleRepository(db)
        self.refresh_repo = RefreshTokenRepository(db)

    @staticmethod
    def _require_host(user: User | None) -> None:
        if user and user.role != "host":
            raise HTTPException(status_code=403, detail="User role is not allowed")

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

    async def login_user(self, user_to_login: UserLogin) -> tuple[TokenPair, User]:
        email = str(user_to_login.email).lower()
        user = await self.user_repo.get_by_email(email)
        if not user:
            existing_user = await self.user_repo.get_existing_by_email(email)
            if existing_user and existing_user.deleted_at is not None:
                raise HTTPException(status_code=403, detail="User is deactivated")
            raise HTTPException(status_code=401, detail="Incorrect email or password")
        self._require_host(user)
        if not verify_password(user_to_login.password, user.password_hash):
            raise HTTPException(status_code=401, detail="Incorrect email or password")
        await self.user_repo.update_last_login(user.id)

        token_pair = create_tokens(user.id, user.token_version)
        await self.refresh_repo.create(
            token_pair.session_id,
            user.id,
            hash_refresh_token(token_pair.refresh_token),
            token_pair.refresh_expires_at,
        )
        return token_pair, user

    async def refresh_tokens(self, refresh_token: str) -> TokenPair:
        user_id, _, session_id = decode_token(refresh_token, expected_type="refresh")
        if not user_id or not session_id:
            raise HTTPException(status_code=401, detail="Invalid refresh token")
        stored_token = await self.refresh_repo.get_by_id(session_id)
        if not stored_token:
            raise HTTPException(status_code=403, detail="Refresh token revoked or unknown")
        hashed_token = hash_refresh_token(refresh_token)
        if stored_token.token_hash != hashed_token or stored_token.user_id != user_id:
            raise HTTPException(status_code=403, detail="Refresh token revoked or reused")
        if stored_token.revoked_at is not None:
            raise HTTPException(status_code=403, detail="Refresh token revoked or reused")
        expires_at = stored_token.expires_at
        if expires_at.tzinfo is None:
            expires_at = expires_at.replace(tzinfo=timezone.utc)
        if expires_at <= datetime.now(timezone.utc):
            await self.refresh_repo.revoke(stored_token)
            raise HTTPException(status_code=401, detail="Refresh token expired")
        user = await self.user_repo.get_by_id(stored_token.user_id)
        if not user:
            raise HTTPException(status_code=404, detail="User not found")
        self._require_host(user)
        await self.refresh_repo.revoke(stored_token)
        token_pair = create_tokens(user.id, user.token_version)
        await self.refresh_repo.create(
            token_pair.session_id,
            user.id,
            hash_refresh_token(token_pair.refresh_token),
            token_pair.refresh_expires_at,
        )
        return token_pair

    async def logout_user(self, user: User, session_id: UUID | None = None) -> None:
        if session_id:
            await self.refresh_repo.revoke_by_id(session_id)
        else:
            await self.refresh_repo.revoke_all_for_user(user.id)
