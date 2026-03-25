from datetime import datetime, timezone
from typing import Tuple
from uuid import UUID

from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from sqlmodel.ext.asyncio.session import AsyncSession

from backend.core.database import get_session
from backend.repositories.refresh_token_repository import RefreshTokenRepository
from backend.repositories.user_repository import UserRepository
from backend.utils.security import decode_token

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/auth/login")


async def _resolve_current_session(token: str, db: AsyncSession) -> Tuple["User", UUID]:
    user_id, token_version, session_id = decode_token(token, "access")
    if not user_id or not session_id:
        raise HTTPException(
            status_code=401,
            detail="Invalid or expired token",
            headers={"WWW-Authenticate": "Bearer"},
        )

    repo = UserRepository(db)
    user = await repo.get_by_id(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    if user.role != "host":
        raise HTTPException(
            status_code=403,
            detail="User role is not allowed",
        )

    if user.token_version != token_version:
        raise HTTPException(
            status_code=401,
            detail="Token has been revoked",
            headers={"WWW-Authenticate": "Bearer"},
        )

    session_repo = RefreshTokenRepository(db)
    session = await session_repo.get_by_id(session_id)
    if not session or session.user_id != user.id:
        raise HTTPException(status_code=401, detail="Session is invalid")
    if session.revoked_at is not None:
        raise HTTPException(status_code=401, detail="Session has been revoked")
    expires_at = session.expires_at
    expires_at = (
        expires_at.replace(tzinfo=timezone.utc) if expires_at.tzinfo is None else expires_at
    )
    if expires_at <= datetime.now(timezone.utc):
        await session_repo.revoke(session)
        raise HTTPException(status_code=401, detail="Session has expired")

    return user, session_id


async def get_current_user(
    token: str = Depends(oauth2_scheme), db: AsyncSession = Depends(get_session)
):
    user, _ = await _resolve_current_session(token, db)
    return user


async def get_current_user_with_session(
    token: str = Depends(oauth2_scheme), db: AsyncSession = Depends(get_session)
):
    return await _resolve_current_session(token, db)
