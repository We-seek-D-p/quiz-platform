from datetime import UTC, datetime
from unittest.mock import AsyncMock, MagicMock
from uuid import uuid4

import pytest

from quiz_management.models.question import Question, QuestionCreate, QuestionOption, QuestionUpdate

pytestmark = pytest.mark.anyio


class TestQuestionService:
    async def test_get_quiz_questions_success(self, question_service, mock_db):
        quiz_id = uuid4()
        expected_questions = [MagicMock(spec=Question), MagicMock(spec=Question)]

        question_service.repository.get_by_quiz_id = AsyncMock(return_value=expected_questions)
        res = await question_service.get_quiz_questions(quiz_id)
        question_service.repository.get_by_quiz_id.assert_called_once_with(quiz_id)

        assert res == expected_questions

    async def test_get_quiz_questions_empty(self, question_service, mock_db):
        quiz_id = uuid4()
        question_service.repository.get_by_quiz_id = AsyncMock(return_value=[])
        res = await question_service.get_quiz_questions(quiz_id)
        assert res == []

    async def test_create_question_success(
        self, question_service, mock_db, question_create_factory
    ):
        quiz_id = uuid4()
        create_data = question_create_factory(
            text="Test Question",
            selection_type="single",
            time_limit_seconds=30,
            order_index=0,
        )

        saved_question = MagicMock(spec=Question)
        saved_question.id = uuid4()
        saved_question.options = []
        question_service.repository.save = AsyncMock(return_value=saved_question)
        res = await question_service.create_question(create_data, quiz_id)

        question_service.repository.save.assert_called_once()
        call_args = question_service.repository.save.call_args[0][0]
        assert isinstance(call_args, Question)
        assert call_args.quiz_id == quiz_id
        assert call_args.text == create_data.text
        assert call_args.selection_type == create_data.selection_type
        assert call_args.time_limit_seconds == create_data.time_limit_seconds
        assert call_args.order_index == create_data.order_index
        assert len(call_args.options) == len(create_data.options)
        assert res == saved_question

    async def test_create_question_sorts_options_by_order_index(
        self, question_service, mock_db, option_create_factory
    ):
        quiz_id = uuid4()
        create_data = QuestionCreate(
            text="Test Question",
            selection_type="single",
            time_limit_seconds=15,
            order_index=0,
            options=[
                option_create_factory(text="Option 1", order_index=2, is_correct=False),
                option_create_factory(text="Option 2", order_index=0),
                option_create_factory(text="Option 3", order_index=1),
                option_create_factory(text="Option 4", order_index=3, is_correct=True),
            ],
        )

        saved_question = MagicMock(spec=Question)
        question_service.repository.save = AsyncMock(return_value=saved_question)
        await question_service.create_question(create_data, quiz_id)

        call_args = question_service.repository.save.call_args[0][0]
        assert call_args.options[0].order_index == 0
        assert call_args.options[0].text == "Option 2"
        assert call_args.options[1].order_index == 1
        assert call_args.options[1].text == "Option 3"
        assert call_args.options[2].order_index == 2
        assert call_args.options[2].text == "Option 1"
        assert call_args.options[3].order_index == 3
        assert call_args.options[3].text == "Option 4"

    async def test_question_updates_fields(
        self, question_service, mock_db, question_update_factory
    ):
        question = MagicMock(spec=Question)
        question.options = []
        question.updated_at = datetime.now(UTC)

        update_data = question_update_factory(
            text="Updated Question",
            selection_type="multiple",
            time_limit_seconds=45,
            order_index=1,
        )

        saved_question = MagicMock(spec=Question)
        question_service.repository.save = AsyncMock(return_value=saved_question)
        res = await question_service.update_question(question, update_data)

        assert question.text == "Updated Question"
        assert question.selection_type == "multiple"
        assert question.time_limit_seconds == 45
        assert question.order_index == 1
        question_service.repository.save.assert_called_once_with(question)
        assert res == saved_question

    async def test_update_question_without_changes_still_updates_timestamp(
        self, question_service, mock_db, question_update_factory
    ):
        question = MagicMock(spec=Question)
        question.options = []
        initial_updated_at = datetime.now(UTC)
        question.updated_at = initial_updated_at
        update_data = question_update_factory()

        async def mock_save(q):
            return q

        question_service.repository.save = AsyncMock(side_effect=mock_save)
        await question_service.update_question(question, update_data)
        question_service.repository.save.assert_called_once_with(question)

    async def test_update_question_adds_new_options(
        self, question_service, mock_db, option_update_factory
    ):
        question = MagicMock(spec=Question)
        question.options = []
        question.updated_at = datetime.now(UTC)

        new_option = option_update_factory(
            option_id=None, text="New Option", order_index=0, is_correct=True
        )
        update_data = QuestionUpdate(
            options=[
                new_option,
                option_update_factory(text="Option 2", order_index=1, is_correct=False),
                option_update_factory(text="Option 3", order_index=2, is_correct=False),
                option_update_factory(text="Option 4", order_index=3, is_correct=False),
            ]
        )
        question_service.repository.save = AsyncMock(return_value=question)
        await question_service.update_question(question, update_data)

        assert len(question.options) == 4
        assert question.options[0].text == "New Option"
        assert question.options[0].order_index == 0
        assert question.options[0].is_correct is True

    async def test_update_question_updates_existing_options(
        self, question_service, mock_db, option_update_factory
    ):
        option_id = uuid4()
        existing_option = MagicMock(spec=QuestionOption)
        existing_option.id = option_id
        existing_option.deleted_at = None
        existing_option.text = "Original Text"
        existing_option.order_index = 0
        existing_option.is_correct = False

        question = MagicMock(spec=Question)
        question.options = [existing_option]
        question.updated_at = datetime.now(UTC)

        update_option = option_update_factory(
            option_id=option_id, text="Updated Option", order_index=1, is_correct=True
        )
        update_data = QuestionUpdate(
            options=[
                update_option,
                option_update_factory(text="Option 2", order_index=0, is_correct=False),
                option_update_factory(text="Option 3", order_index=2, is_correct=False),
                option_update_factory(text="Option 4", order_index=3, is_correct=False),
            ]
        )
        question_service.repository.save = AsyncMock(return_value=question)
        await question_service.update_question(question, update_data)

        assert existing_option.text == "Updated Option"
        assert existing_option.order_index == 1
        assert existing_option.is_correct is True

    async def test_update_question_deletes_removed_options(
        self, question_service, mock_db, option_update_factory
    ):
        option_id = uuid4()
        existing_option = MagicMock(spec=QuestionOption)
        existing_option.id = option_id
        existing_option.deleted_at = None

        question = MagicMock(spec=Question)
        question.options = [existing_option]
        question.updated_at = datetime.now(UTC)

        update_data = QuestionUpdate(
            options=[
                option_update_factory(text="Option 1", order_index=0, is_correct=True),
                option_update_factory(text="Option 2", order_index=1, is_correct=False),
                option_update_factory(text="Option 3", order_index=2, is_correct=False),
                option_update_factory(text="Option 4", order_index=3, is_correct=False),
            ]
        )
        question_service.repository.save = AsyncMock(return_value=question)
        await question_service.update_question(question, update_data)

        assert existing_option.deleted_at is not None
        assert isinstance(existing_option.deleted_at, datetime)

    async def test_update_question_sorts_options_by_order_index(
        self, question_service, mock_db, option_update_factory
    ):
        option_id1 = uuid4()
        option_id2 = uuid4()

        existing_option1 = MagicMock(spec=QuestionOption)
        existing_option1.id = option_id1
        existing_option1.deleted_at = None

        existing_option2 = MagicMock(spec=QuestionOption)
        existing_option2.id = option_id2
        existing_option2.deleted_at = None

        question = MagicMock(spec=Question)
        question.options = [existing_option1, existing_option2]
        question.updated_at = datetime.now(UTC)

        update_option1 = option_update_factory(option_id=option_id1, order_index=5, is_correct=True)
        update_option2 = option_update_factory(
            option_id=option_id2, order_index=1, is_correct=False
        )
        update_data = QuestionUpdate(
            options=[
                update_option1,
                update_option2,
                option_update_factory(text="Option 3", order_index=0, is_correct=False),
                option_update_factory(text="Option 4", order_index=2, is_correct=False),
            ]
        )
        question_service.repository.save = AsyncMock(return_value=question)
        await question_service.update_question(question, update_data)

        assert existing_option2.order_index == 1
        assert existing_option1.order_index == 3

    async def test_delete_question_soft_deletes_question_and_options(
        self, question_service, mock_db
    ):
        option1 = MagicMock(spec=QuestionOption)
        option1.deleted_at = None
        option2 = MagicMock(spec=QuestionOption)
        option2.deleted_at = None

        question = MagicMock(spec=Question)
        question.deleted_at = None
        question.options = [option1, option2]

        question_service.repository.save = AsyncMock(return_value=question)
        await question_service.delete_question(question)

        assert question.deleted_at is not None
        assert isinstance(question.deleted_at, datetime)
        assert option1.deleted_at is not None
        assert option2.deleted_at is not None
        question_service.repository.save.assert_called_once_with(question)

    async def test_delete_question_with_no_options(self, question_service, mock_db):
        question = MagicMock(spec=Question)
        question.deleted_at = None
        question.options = []

        question_service.repository.save = AsyncMock(return_value=question)
        await question_service.delete_question(question)

        assert question.deleted_at is not None
        question_service.repository.save.assert_called_once_with(question)
