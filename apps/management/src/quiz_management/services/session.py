from repositories.session_repositories import SessionRepository, TSession
from sqlmodel.ext.asyncio.session import AsyncSession


class SessionService:
    def __init__(self, db: AsyncSession):
        self.repository = SessionRepository(db)

    async def create_session(self, data: TSession) -> None:
        await self.repository.save_session(data)
