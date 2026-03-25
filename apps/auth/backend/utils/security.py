from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from hashlib import sha256
from uuid import UUID, uuid7

import jwt

from pwdlib import PasswordHash

from backend.core.config import settings


password_hash = PasswordHash.recommended()


@dataclass(frozen=True)
class TokenPair:
    access_token: str
    refresh_token: str
    token_type: str
    access_expires_in: int
    refresh_expires_at: datetime
    refresh_expires_in: int
    session_id: UUID


def hash_password(password: str) -> str:
    return password_hash.hash(password)


def verify_password(plain: str, hashed: str) -> bool:
    return password_hash.verify(plain, hashed)


def hash_refresh_token(refresh_token: str) -> str:
    return sha256(refresh_token.encode("utf-8")).hexdigest()


def create_tokens(user_id: UUID, token_version: int, session_id: UUID | None = None) -> TokenPair:
    now = datetime.now(timezone.utc)
    access_expires_delta = timedelta(minutes=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES)
    refresh_expires_delta = timedelta(days=settings.JWT_REFRESH_TOKEN_EXPIRE_DAYS)
    session = session_id or uuid7()

    access_payload = {
        "sub": str(user_id),
        "ver": token_version,
        "exp": now + access_expires_delta,
        "type": "access",
        "sid": str(session),
    }
    access_token = jwt.encode(
        access_payload, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM
    )

    refresh_expires_at = now + refresh_expires_delta
    refresh_payload = {
        "sub": str(user_id),
        "exp": refresh_expires_at,
        "type": "refresh",
        "sid": str(session),
    }
    refresh_token = jwt.encode(
        refresh_payload, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM
    )

    return TokenPair(
        access_token=access_token,
        refresh_token=refresh_token,
        token_type="Bearer",
        access_expires_in=int(access_expires_delta.total_seconds()),
        refresh_expires_at=refresh_expires_at,
        refresh_expires_in=int(refresh_expires_delta.total_seconds()),
        session_id=session,
    )


def decode_token(
    token: str, expected_type: str = "access"
) -> tuple[UUID | None, int | None, UUID | None]:
    try:
        payload = jwt.decode(token, settings.JWT_SECRET_KEY, algorithms=[settings.JWT_ALGORITHM])
        if payload.get("type") != expected_type:
            return None, None, None
        user_id = payload.get("sub")
        token_version = payload.get("ver", 0)
        session_id = payload.get("sid")
        if user_id is None:
            return None, None, None
        try:
            session_uuid = UUID(session_id) if session_id else None
            return UUID(user_id), token_version, session_uuid
        except ValueError, TypeError:
            return None, None, None
    except jwt.ExpiredSignatureError, jwt.InvalidTokenError:
        return None, None, None
