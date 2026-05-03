from uuid import UUID

from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.core.exceptions import ServiceException
from quiz_management.models.quiz import Quiz
from quiz_management.models.session import GameSession, SessionStatus
from quiz_management.repositories.session_repositories import SessionRepository
from quiz_management.services.session_client import SessionServiceClient


class SessionService:
    def __init__(self, db: AsyncSession, session_client: SessionServiceClient):
        self.repository = SessionRepository(db)
        self.client = session_client

    async def create_session(self, quiz: Quiz, user_id: UUID, idempotency_key: str) -> GameSession:
        if not quiz.questions:
            raise ServiceException(
                status_code=400,
                code="quiz_not_ready",
                message="Cannot start session for quiz without questions",
            )

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
            raise ServiceException(
                status_code=424,
                code="session_provider_error",
                message="Go session service failed to initialize runtime",
            )
        except Exception:
            new_session.status = SessionStatus.INIT_FAILED
            await self.repository.save_session(new_session)
            await self.client.delete_session(new_session.id)
            raise ServiceException(
                status_code=503,
                code="session_provider_unavailable",
                message="Go session service is not responding",
            ) from None

    async def get_bootstrap_data(self, session_id: UUID) -> GameSession:
        session = await self.repository.get_session_with_quiz(session_id)

        if not session:
            raise ServiceException(404, "session_not_found", "Session not found")

        if session.status == SessionStatus.FINISHED:
            raise ServiceException(409, "already_finished", "Session already finished")

        if not session.quiz:
            raise ServiceException(404, "quiz_not_found", "Linked quiz not found")

        return session
