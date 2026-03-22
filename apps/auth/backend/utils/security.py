from datetime import datetime, timedelta
import jwt
from passlib.context import CryptContext
from backend.core.config import settings

from pwdlib import PasswordHash

password_hash = PasswordHash.recommended()


def hash_password(password: str) -> str:
    return password_hash.hash(password)


def verify_password(plain: str, hashed: str) -> bool:
    return password_hash.verify(plain, hashed)


def create_tokens(user_id: int, token_version: int) -> dict[str, str]:
    now = datetime.now()

    access_payload = {
        "sub": str(user_id),
        "ver": token_version,
        "exp": now + timedelta(minutes=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES),
        "type": "access"
    }
    access_token = jwt.encode(access_payload, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM)

    refresh_payload = {
        "sub": str(user_id),
        "exp": now + timedelta(minutes=settings.JWT_REFRESH_TOKEN_EXPIRE_DAYS),
        "type": "refresh"
    }
    refresh_token = jwt.encode(refresh_payload, settings.JWT_SECRET_KEY, algorithm=settings.JWT_ALGORITHM)

    return {"access_token": access_token, "refresh_token": refresh_token, "token_type": "bearer"}


def decode_token(token: str, expected_type: str = "access") -> tuple[int | None, int | None]:
    try:
        payload = jwt.decode(
            token,
            settings.JWT_SECRET_KEY,
            algorithms=[settings.JWT_ALGORITHM]
        )
        if payload.get("type") != expected_type:
            return None, None
        user_id = payload.get("sub")
        token_version = payload.get("ver", 0)
        if user_id is None:
            return None, None
        return int(user_id) if user_id else None, token_version
    except (jwt.ExpiredSignatureError, jwt.InvalidTokenError):
        return None, None
