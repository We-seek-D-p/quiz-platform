from datetime import UTC, datetime
from unittest.mock import AsyncMock, MagicMock
from uuid import uuid7

import httpx
import pytest

from quiz_management.core.exceptions import ServiceException
from quiz_management.models.session import GameSession, SessionStatus

pytestmark = pytest.mark.anyio


class TestSessionService:
    async def test_create_session_success(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        mock_quiz.questions = [MagicMock()]
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        session = await session_service.create_session(mock_quiz, uuid7(), "key")

        assert session.status == SessionStatus.LOBBY
        assert session.room_code == "123456"
        assert session_service.repository.save_session.call_count == 2

    async def test_create_session_rejects_quiz_without_questions(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        mock_quiz.questions = []

        with pytest.raises(ServiceException) as exc:
            await session_service.create_session(mock_quiz, uuid7(), "key")

        assert exc.value.status_code == 400
        assert exc.value.detail["code"] == "quiz_not_ready"

    async def test_create_game_session(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        mock_quiz.questions = [MagicMock()]
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        await session_service.create_session(mock_quiz, uuid7(), "key")

        assert session_service.repository.save_session.call_count >= 1

    async def test_create_session_participant(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        mock_quiz.questions = [MagicMock()]
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        await session_service.create_session(mock_quiz, uuid7(), "key")
        assert session_service.repository.save_session.call_count >= 1

    async def test_create_session_without_refresh(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        mock_quiz.questions = [MagicMock()]
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        await session_service.create_session(mock_quiz, uuid7(), "key")
        assert session_service.repository.save_session.call_count >= 1

    async def test_create_session_multiple_calls(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        mock_quiz.questions = [MagicMock()]
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        await session_service.create_session(mock_quiz, uuid7(), "key-1")
        await session_service.create_session(mock_quiz, uuid7(), "key-2")

        assert session_service.repository.save_session.call_count >= 2

    async def test_create_session_maps_provider_payload_error_to_failed_dependency(
        self, session_service, mock_db
    ):
        mock_quiz = MagicMock(id=uuid7())
        mock_quiz.questions = [MagicMock()]
        session_service.repository.save_session = AsyncMock()
        session_service.client.delete_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(
                status_code=409,
                json=lambda: {"code": "duplicate_session", "message": "Duplicate session"},
            )
        )

        with pytest.raises(ServiceException) as exc:
            await session_service.create_session(mock_quiz, uuid7(), "key")

        assert exc.value.status_code == 424
        assert exc.value.detail == {"code": "duplicate_session", "message": "Duplicate session"}
        session_service.client.delete_session.assert_called_once()

    async def test_create_session_maps_provider_unavailable_error(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        mock_quiz.questions = [MagicMock()]
        session_service.repository.save_session = AsyncMock()
        session_service.client.delete_session = AsyncMock()
        session_service.client.init_session = AsyncMock(return_value=MagicMock(status_code=503))

        with pytest.raises(ServiceException) as exc:
            await session_service.create_session(mock_quiz, uuid7(), "key")

        assert exc.value.status_code == 503
        assert exc.value.detail["code"] == "session_provider_unavailable"
        session_service.client.delete_session.assert_called_once()

    async def test_create_session_reconciles_runtime_after_http_error(
        self, session_service, mock_db
    ):
        mock_quiz = MagicMock(id=uuid7())
        mock_quiz.questions = [MagicMock()]
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(side_effect=httpx.ConnectError("boom"))
        session_service.client.get_session = AsyncMock(
            return_value=MagicMock(status_code=200, json=lambda: {"room_code": "654321"})
        )

        session = await session_service.create_session(mock_quiz, uuid7(), "key")

        assert session.status == SessionStatus.LOBBY
        assert session.room_code == "654321"
        assert session_service.repository.save_session.call_count == 2

    async def test_update_session_status_rejects_blank_event_id(self, session_service, mock_db):
        data = MagicMock(event_id="   ", status=SessionStatus.LOBBY, started_at=None)

        with pytest.raises(ServiceException) as exc:
            await session_service.update_session_status(uuid7(), data)

        assert exc.value.status_code == 400
        assert exc.value.detail["code"] == "invalid_payload"

    async def test_update_session_status_rejects_invalid_transition(self, session_service, mock_db):
        session = GameSession(quiz_id=uuid7(), host_id=uuid7(), status=SessionStatus.LOBBY)
        session_service.repository.get_session_by_id = AsyncMock(return_value=session)
        data = MagicMock(event_id="event-1", status=SessionStatus.FINISHED, started_at=None)

        with pytest.raises(ServiceException) as exc:
            await session_service.update_session_status(session.id, data)

        assert exc.value.status_code == 409
        assert exc.value.detail["code"] == "invalid_state_transition"

    async def test_update_session_status_sets_started_at_as_naive_utc(
        self, session_service, mock_db
    ):
        session = GameSession(quiz_id=uuid7(), host_id=uuid7(), status=SessionStatus.LOBBY)
        started_at = datetime(2026, 1, 2, 3, 4, 5, tzinfo=UTC)
        session_service.repository.get_session_by_id = AsyncMock(return_value=session)
        session_service.repository.save_session = AsyncMock()
        data = MagicMock(
            event_id="event-1", status=SessionStatus.IN_PROGRESS, started_at=started_at
        )

        await session_service.update_session_status(session.id, data)

        assert session.status == SessionStatus.IN_PROGRESS
        assert session.started_at == datetime(2026, 1, 2, 3, 4, 5)
        session_service.repository.save_session.assert_called_once_with(session)

    async def test_finalize_session_persists_results(self, session_service, mock_db):
        session = GameSession(quiz_id=uuid7(), host_id=uuid7(), status=SessionStatus.IN_PROGRESS)
        finished_at = datetime(2026, 1, 2, 3, 4, 5, tzinfo=UTC)
        participant = MagicMock(nickname="Alice", score=42, rank=1)
        data = MagicMock(event_id="event-1", finished_at=finished_at, participants=[participant])
        session_service.repository.get_session_with_quiz = AsyncMock(return_value=session)
        session_service.repository.save_results = AsyncMock()

        await session_service.finalize_session(session.id, data)

        saved_session, participants = session_service.repository.save_results.call_args.args
        assert saved_session == session
        assert session.status == SessionStatus.FINISHED
        assert session.finished_at == datetime(2026, 1, 2, 3, 4, 5)
        assert len(participants) == 1
        assert participants[0].player_nickname == "Alice"
        assert participants[0].score == 42
        assert participants[0].rank == 1
