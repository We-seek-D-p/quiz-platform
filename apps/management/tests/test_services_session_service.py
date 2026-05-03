from unittest.mock import AsyncMock, MagicMock
from uuid import uuid7

import pytest

pytestmark = pytest.mark.anyio


class TestSessionService:
    async def test_create_session_success(self, session_service, mock_db):
        mock_quiz = MagicMock()
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        await session_service.create_session(mock_quiz, uuid7(), "key")

        assert session_service.repository.save_session.call_count >= 1

    async def test_create_game_session(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        await session_service.create_session(mock_quiz, uuid7(), "key")

        assert session_service.repository.save_session.call_count >= 1

    async def test_create_session_participant(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        await session_service.create_session(mock_quiz, uuid7(), "key")
        assert session_service.repository.save_session.call_count >= 1

    async def test_create_session_without_refresh(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        await session_service.create_session(mock_quiz, uuid7(), "key")
        assert session_service.repository.save_session.call_count >= 1

    async def test_create_session_multiple_calls(self, session_service, mock_db):
        mock_quiz = MagicMock(id=uuid7())
        session_service.repository.save_session = AsyncMock()
        session_service.client.init_session = AsyncMock(
            return_value=MagicMock(status_code=201, json=lambda: {"room_code": "123456"})
        )

        await session_service.create_session(mock_quiz, uuid7(), "key-1")
        await session_service.create_session(mock_quiz, uuid7(), "key-2")

        assert session_service.repository.save_session.call_count >= 2
