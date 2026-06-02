from unittest.mock import AsyncMock, MagicMock
from uuid import uuid4

import pytest

from quiz_management.repositories.quiz_repository import QuizRepository

pytestmark = pytest.mark.anyio


class TestQuizRepository:
    async def test_get_by_owner_id_filters_deleted_quizzes_and_orders_by_created_at(self):
        owner_id = uuid4()
        expected = [MagicMock(), MagicMock()]
        result = MagicMock()
        result.all.return_value = expected
        db = MagicMock()
        db.exec = AsyncMock(return_value=result)
        repository = QuizRepository(db)

        res = await repository.get_by_owner_id(owner_id)

        statement = db.exec.call_args.args[0]
        compiled = str(statement.compile(compile_kwargs={"literal_binds": False}))
        assert "quizzes.owner_id" in compiled
        assert "quizzes.deleted_at IS NULL" in compiled
        assert "ORDER BY" in compiled
        assert "quizzes.created_at DESC" in compiled
        assert res == expected
