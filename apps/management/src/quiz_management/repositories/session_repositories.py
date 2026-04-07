from typing import TypeVar

from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.models.session import GameSession, SessionParticipant

TSession = TypeVar("TSession", GameSession, SessionParticipant)


class SessionRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def save_session(self, session: TSession) -> None:
        self.db.add(session)
        await self.db.commit()
        await self.db.refresh(session)
