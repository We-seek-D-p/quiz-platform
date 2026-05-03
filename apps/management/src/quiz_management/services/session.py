from uuid import UUID

from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.repositories.session_repositories import SessionRepository
from quiz_management.services.session_client import SessionServiceClient


class SessionService:
    def __init__(self, db: AsyncSession):
        self.repository = SessionRepository(db)
        self.client = SessionServiceClient()

    async def create_session(self, user_id: UUID, quiz_id: UUID, idempotency_key: str):
        pass
