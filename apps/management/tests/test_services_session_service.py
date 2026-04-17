from unittest.mock import AsyncMock, MagicMock

import pytest

from quiz_management.models.session import GameSession, SessionParticipant

pytestmark = pytest.mark.anyio


class TestSessionService:
    async def test_create_session_success(self, session_service, mock_db):
        mock_session_data = MagicMock()
        session_service.repository.save_session = AsyncMock()

        await session_service.create_session(mock_session_data)

        session_service.repository.save_session.assert_called_once_with(mock_session_data)

    async def test_create_game_session(self, session_service, mock_db):
        game_session = GameSession(
            quiz_id=MagicMock(),
            room_code="ABC123",
            host_id=MagicMock(),
            status="waiting",
        )

        session_service.repository.save_session = AsyncMock()
        await session_service.create_session(game_session)
        session_service.repository.save_session.assert_called_once_with(game_session)

    async def test_create_session_participant(self, session_service, mock_db):
        participant = SessionParticipant(
            session_id=MagicMock(),
            player_nickname="Player1",
            score=0,
        )

        session_service.repository.save_session = AsyncMock()
        await session_service.create_session(participant)
        session_service.repository.save_session.assert_called_once_with(participant)

    async def test_create_session_without_refresh(self, session_service, mock_db):
        mock_session_data = MagicMock()
        session_service.repository.save_session = AsyncMock()
        await session_service.create_session(mock_session_data)
        assert session_service.repository.save_session.call_count == 1

    async def test_create_session_multiple_calls(self, session_service, mock_db):
        session1 = MagicMock()
        session2 = MagicMock()

        session_service.repository.save_session = AsyncMock()

        await session_service.create_session(session1)
        await session_service.create_session(session2)

        assert session_service.repository.save_session.call_count == 2
        session_service.repository.save_session.assert_any_call(session1)
        session_service.repository.save_session.assert_any_call(session2)
