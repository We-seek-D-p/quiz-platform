from datetime import UTC, datetime
from uuid import UUID

import httpx
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.core.exceptions import ServiceException
from quiz_management.models.quiz import Quiz
from quiz_management.models.session import (
    GameSession,
    SessionParticipant,
    SessionResultsUpdate,
    SessionStatus,
    SessionStatusUpdate,
)
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
        except httpx.HTTPError:
            reconcile_response = await self._reconcile_session_runtime(new_session.id)
            if reconcile_response and reconcile_response.status_code == 200:
                return await self._apply_initialized_runtime(new_session, reconcile_response)

            await self._mark_init_failed_and_compensate(new_session)
            if reconcile_response and reconcile_response.status_code == 404:
                raise ServiceException(
                    status_code=424,
                    code="session_provider_error",
                    message="Go session service runtime was not found during reconcile",
                ) from None

            raise ServiceException(
                status_code=503,
                code="session_provider_unavailable",
                message="Go session service is not responding",
            ) from None

        if response.status_code in (200, 201):
            return await self._apply_initialized_runtime(new_session, response)

        await self._mark_init_failed_and_compensate(new_session)
        raise self._map_init_error_response(response)

    async def _apply_initialized_runtime(
        self, session: GameSession, response: httpx.Response
    ) -> GameSession:
        data = response.json()
        session.room_code = data["room_code"]
        session.status = SessionStatus.LOBBY
        await self.repository.save_session(session)
        return session

    async def _mark_init_failed_and_compensate(self, session: GameSession) -> None:
        session.status = SessionStatus.INIT_FAILED
        await self.repository.save_session(session)
        try:
            await self.client.delete_session(session.id)
        except httpx.HTTPError:
            return

    async def _reconcile_session_runtime(self, session_id: UUID) -> httpx.Response | None:
        try:
            return await self.client.get_session(session_id)
        except httpx.HTTPError:
            return None

    def _map_init_error_response(self, response: httpx.Response) -> ServiceException:
        code = "session_provider_error"
        message = "Go session service failed to initialize runtime"

        try:
            payload = response.json()
            if isinstance(payload, dict):
                upstream_code = payload.get("code")
                if isinstance(upstream_code, str) and upstream_code.strip():
                    code = upstream_code
                upstream_message = payload.get("message")
                if isinstance(upstream_message, str) and upstream_message.strip():
                    message = upstream_message
        except ValueError:
            pass

        status_code = response.status_code
        if status_code in (400, 404, 409, 424):
            return ServiceException(status_code=424, code=code, message=message)
        if status_code in (401, 403, 500, 502, 503, 504):
            return ServiceException(
                status_code=503,
                code="session_provider_unavailable",
                message="Go session service is not responding",
            )

        return ServiceException(status_code=424, code=code, message=message)

    async def get_bootstrap_data(self, session_id: UUID) -> GameSession:
        session = await self.repository.get_session_with_quiz(session_id)

        if not session:
            raise ServiceException(404, "session_not_found", "Session not found")

        if session.status == SessionStatus.FINISHED:
            raise ServiceException(409, "already_finished", "Session already finished")

        if not session.quiz:
            raise ServiceException(404, "quiz_not_found", "Linked quiz not found")

        return session

    async def update_session_status(self, session_id: UUID, data: SessionStatusUpdate) -> None:
        if not data.event_id.strip():
            raise ServiceException(400, "invalid_payload", "event_id is required")

        session = await self.repository.get_session_by_id(session_id)
        if not session:
            raise ServiceException(404, "session_not_found", "Session not found")

        if session.status == data.status:
            await self.repository.save_session(session)
            return

        if session.status == SessionStatus.FINISHED:
            raise ServiceException(
                409, "already_finished", "Cannot change status of a finished session"
            )

        if not self._is_valid_status_transition(session.status, data.status):
            raise ServiceException(
                409,
                "invalid_state_transition",
                f"Cannot transition status from {session.status} to {data.status}",
            )

        session.status = data.status
        if data.status == SessionStatus.IN_PROGRESS and data.started_at:
            session.started_at = self._to_naive_utc(data.started_at)

        await self.repository.save_session(session)

    async def finalize_session(self, session_id: UUID, data: SessionResultsUpdate) -> None:
        if not data.event_id.strip():
            raise ServiceException(400, "invalid_payload", "event_id is required")

        session = await self.repository.get_session_with_quiz(session_id)
        if not session:
            raise ServiceException(404, "session_not_found", "Session not found")

        if session.status == SessionStatus.FINISHED:
            raise ServiceException(409, "already_finished", "Session already finished")

        if not self._is_valid_results_transition(session.status):
            raise ServiceException(
                409,
                "invalid_state_transition",
                f"Cannot finalize session from status {session.status}",
            )

        participants = []
        for p_data in data.participants:
            participant = SessionParticipant(
                session_id=session_id,
                player_nickname=p_data.nickname,
                score=p_data.score,
                rank=p_data.rank,
            )
            participants.append(participant)

        session.status = SessionStatus.FINISHED
        session.finished_at = self._to_naive_utc(data.finished_at)

        await self.repository.save_results(session, participants)

    @staticmethod
    def _is_valid_status_transition(current: SessionStatus, target: SessionStatus) -> bool:
        allowed_transitions = {
            SessionStatus.INITIALIZING: {SessionStatus.LOBBY, SessionStatus.INIT_FAILED},
            SessionStatus.LOBBY: {SessionStatus.IN_PROGRESS},
            SessionStatus.IN_PROGRESS: {SessionStatus.FINISHED},
            SessionStatus.FINISHED: set(),
            SessionStatus.INIT_FAILED: set(),
        }

        return target in allowed_transitions.get(current, set())

    @staticmethod
    def _is_valid_results_transition(current: SessionStatus) -> bool:
        return current in {SessionStatus.LOBBY, SessionStatus.IN_PROGRESS}

    @staticmethod
    def _to_naive_utc(value: datetime | None) -> datetime | None:
        if value is None:
            return None
        if value.tzinfo is None:
            return value
        return value.astimezone(UTC).replace(tzinfo=None)
