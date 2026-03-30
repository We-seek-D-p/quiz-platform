from datetime import UTC, datetime
from typing import Any, cast
from uuid import UUID

from apps.auth.src.quiz_auth.models.users import RefreshToken
from sqlmodel import select, update
from sqlmodel.ext.asyncio.session import AsyncSession

REFRESH_TOKEN_TABLE: Any = cast(Any, RefreshToken).__table__
_REVOKED_AT_COLUMN = REFRESH_TOKEN_TABLE.c.revoked_at
_USER_ID_COLUMN = REFRESH_TOKEN_TABLE.c.user_id


class RefreshTokenRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def create(
        self, token_id: UUID, user_id: UUID, token_hash: str, expires_at: datetime
    ) -> RefreshToken:
        refresh_token = RefreshToken(
            id=token_id,
            user_id=user_id,
            token_hash=token_hash,
            expires_at=expires_at,
        )
        self.db.add(refresh_token)
        await self.db.commit()
        await self.db.refresh(refresh_token)
        return refresh_token

    async def get_by_id(self, token_id: UUID) -> RefreshToken | None:
        result = await self.db.exec(select(RefreshToken).where(RefreshToken.id == token_id))
        return result.first()

    async def revoke(self, token: RefreshToken) -> None:
        if token.revoked_at is not None:
            return
        token.revoked_at = datetime.now(UTC)
        self.db.add(token)
        await self.db.commit()

    async def revoke_by_id(self, token_id: UUID) -> None:
        token = await self.get_by_id(token_id)
        if not token:
            return
        await self.revoke(token)

    async def revoke_all_for_user(self, user_id: UUID) -> None:
        await self.db.exec(
            update(RefreshToken)
            .where(_USER_ID_COLUMN == user_id, _REVOKED_AT_COLUMN.is_(None))
            .values(revoked_at=datetime.now(UTC))
        )
        await self.db.commit()
