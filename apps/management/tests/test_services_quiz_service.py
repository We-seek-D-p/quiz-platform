from datetime import UTC, datetime
from unittest.mock import AsyncMock
from uuid import uuid4

import pytest

from quiz_management.models.quiz import Quiz

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

    async def test_update_quiz_updates_title_and_description(
        self, quiz_service, mock_db, quiz_update_factory
    ):
        quiz = Quiz(title="Original Title", description="Original Description", owner_id=uuid4())

        update_data = quiz_update_factory(title="Updated Title", description="Updated Description")

        saved_quiz = Quiz(
            title=update_data.title, description=update_data.description, owner_id=quiz.owner_id
        )

        quiz_service.repository.save = AsyncMock(return_value=saved_quiz)
        res = await quiz_service.update_quiz(quiz, update_data)

        assert quiz.title == "Updated Title"
        assert quiz.description == "Updated Description"
        assert quiz.updated_at is not None
        quiz_service.repository.save.assert_called_once()
        assert res == saved_quiz

    async def test_update_quiz_updates_only_title_sets_description_to_none(
        self, quiz_service, mock_db, quiz_update_factory
    ):
        quiz = Quiz(title="Original Title", description="Original Description", owner_id=uuid4())

        update_data = quiz_update_factory(title="New Title Only")
        saved_quiz = Quiz(title="New Title Only", description=None, owner_id=quiz.owner_id)

        quiz_service.repository.save = AsyncMock(return_value=saved_quiz)
        await quiz_service.update_quiz(quiz, update_data)

        assert quiz.title == "New Title Only"
        assert quiz.description is None

    async def test_update_quiz_updates_only_description_sets_title_to_none(
        self, quiz_service, mock_db, quiz_update_factory
    ):
        quiz = Quiz(title="Original Title", description="Original Description", owner_id=uuid4())

        update_data = quiz_update_factory(description="New Description Only")
        saved_quiz = Quiz(title=None, description="New Description Only", owner_id=quiz.owner_id)

        quiz_service.repository.save = AsyncMock(return_value=saved_quiz)
        await quiz_service.update_quiz(quiz, update_data)

        assert quiz.title is None
        assert quiz.description == "New Description Only"

    async def test_update_quiz_with_no_changes_sets_both_to_none(
        self, quiz_service, mock_db, quiz_update_factory
    ):
        quiz = Quiz(title="Original Title", description="Original Description", owner_id=uuid4())

        original_updated_at = quiz.updated_at
        update_data = quiz_update_factory()
        quiz_service.repository.save = AsyncMock(return_value=quiz)
        await quiz_service.update_quiz(quiz, update_data)

        assert quiz.title is None
        assert quiz.description is None
        assert quiz.updated_at is not None
        assert quiz.updated_at != original_updated_at
        quiz_service.repository.save.assert_called_once()

    async def test_update_quiz_with_empty_strings(self, quiz_service, mock_db, quiz_update_factory):
        quiz = Quiz(title="Original Title", description="Original Description", owner_id=uuid4())

        update_data = quiz_update_factory(title="", description="")
        saved_quiz = Quiz(title="", description="", owner_id=quiz.owner_id)
        quiz_service.repository.save = AsyncMock(return_value=saved_quiz)
        await quiz_service.update_quiz(quiz, update_data)

        assert quiz.title == ""
        assert quiz.description == ""
        assert quiz.updated_at is not None

    async def test_delete_quiz_soft_deletes(self, quiz_service, mock_db):
        quiz = Quiz(title="Test Quiz", description="Test Description", owner_id=uuid4())

        quiz.deleted_at = None
        quiz_service.repository.save = AsyncMock(return_value=quiz)
        await quiz_service.delete_quiz(quiz)

        assert quiz.deleted_at is not None
        assert isinstance(quiz.deleted_at, datetime)
        quiz_service.repository.save.assert_called_once_with(quiz)

    async def test_delete_already_deleted_quiz(self, quiz_service, mock_db):
        previous_deleted = datetime.now(UTC).replace(tzinfo=None)
        quiz = Quiz(title="Test Quiz", description="Test Description", owner_id=uuid4())

        quiz.deleted_at = previous_deleted
        quiz_service.repository.save = AsyncMock(return_value=quiz)
        await quiz_service.delete_quiz(quiz)

        assert quiz.deleted_at != previous_deleted
        assert quiz.deleted_at > previous_deleted
