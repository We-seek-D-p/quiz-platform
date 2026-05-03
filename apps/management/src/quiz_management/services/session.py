from uuid import UUID

import httpx
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.models.quiz import Quiz
from quiz_management.models.session import GameSession, SessionStatus
from quiz_management.repositories.session_repositories import SessionRepository
from quiz_management.services.session_client import SessionServiceClient


class SessionService:
    def __init__(self, db: AsyncSession):
        self.repository = SessionRepository(db)
        self.client = SessionServiceClient()

    async def create_session(self, quiz: Quiz, user_id: UUID, idempotency_key: str) -> GameSession:
        new_session = GameSession(
            quiz_id=quiz.id, host_id=user_id, status=SessionStatus.INITIALIZING
        )
        await self.repository.save_session(new_session)

        try:
            response = await self.client.init_session(
                session_id=new_session.id,
                quiz_id=quiz.id,
                host_id=user_id,
                idempotency_key=idempotency_key,
            )

            if response.status_code in (200, 201):
                data = response.json()
                new_session.room_code = data["room_code"]
                new_session.status = SessionStatus.LOBBY
                await self.repository.save_session(new_session)
                return new_session

            new_session.status = SessionStatus.INIT_FAILED
            await self.repository.save_session(new_session)
            await self.client.delete_session(new_session.id)
            raise httpx.HTTPStatusError(
                "Go service returned error", request=response.request, response=response
            )

        except Exception:
            new_session.status = SessionStatus.INIT_FAILED
            await self.repository.save_session(new_session)
            await self.client.delete_session(new_session.id)
            raise
