from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.models.session import GameSession, SessionParticipant


class SessionRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def create_session(self, session: GameSession | SessionParticipant) -> GameSession:
        self.db.add(session)
        await self.db.commit()
        await self.db.refresh(session)
        return session
