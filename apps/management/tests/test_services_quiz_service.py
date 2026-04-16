from datetime import UTC, datetime
from unittest.mock import MagicMock, AsyncMock
from uuid import uuid4

import pytest

from quiz_management.models.quiz import Quiz, QuizCreate, QuizUpdate
from quiz_management.services.quiz import QuizService


pytestmark = pytest.mark.anyio


class TestQuizService:
    async def test_create_quiz_success(self, quiz_service, mock_db, quiz_create_factory):
        user_id = uuid4()
        create_data = quiz_create_factory(title="My Quiz", description="My Description")

        saved_quiz = Quiz(
            title=create_data.title,
            description=create_data.description,
            owner_id=user_id,
        )

        quiz_service.repository.save = AsyncMock(return_value=saved_quiz)
        res = await quiz_service.create_quiz(user_id, create_data)

        quiz_service.repository.save.assert_called_once()
        call_args = quiz_service.repository.save.call_args[0][0]
        assert call_args.title == create_data.title
        assert call_args.description == create_data.description
        assert call_args.owner_id == user_id
        assert res == saved_quiz

    async def test_create_quiz_with_minimal_data(self, quiz_service, mock_db, quiz_create_factory):
        user_id = uuid4()
        create_data = quiz_create_factory(title="MinimalQuiz", description="")
        saved_quiz = Quiz(
            title=create_data.title,
            description=create_data.description,
            owner_id=user_id,
        )
        quiz_service.repository.save = AsyncMock(return_value=saved_quiz)
        res = await quiz_service.create_quiz(user_id, create_data)

        call_args = quiz_service.repository.save.call_args[0][0]
        assert call_args.title == "MinimalQuiz"
        assert call_args.description == ""
        assert res == saved_quiz
