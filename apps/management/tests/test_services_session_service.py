from unittest.mock import AsyncMock, MagicMock
import pytest
from quiz_management.services.session import SessionService
from quiz_management.models.session import GameSession

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
