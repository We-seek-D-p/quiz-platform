from datetime import UTC, datetime
from unittest.mock import MagicMock, AsyncMock
from uuid import uuid4

import pytest

from quiz_management.models.question import Question
from quiz_management.services.question import QuestionService

pytestmark = pytest.mark.anyio


class TestQuestionService:
    async def test_get_quiz_questions_success(self, question_service, mock_db):
        quiz_id = uuid4()
        expected_questions = [MagicMock(spec=Question), MagicMock(spec=Question)]

        question_service.repository.get_by_quiz_id = AsyncMock(return_value=expected_questions)
        res = await question_service.get_quiz_questions(quiz_id)
        question_service.repository.get_by_quiz_id.assert_called_once_with(quiz_id)

        assert res == expected_questions
