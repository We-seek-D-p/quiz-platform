import sys
from unittest.mock import MagicMock

mock_settings = MagicMock()
mock_settings.database_url = "sqlite:///:memory:"
mock_settings.session_provider_url = "http://localhost:8080"

sys.modules["quiz_management.core.config"] = MagicMock()
sys.modules["quiz_management.core.config"].settings = mock_settings

from collections.abc import Callable  # noqa: E402
from typing import Any  # noqa: E402
from unittest.mock import AsyncMock  # noqa: E402
from uuid import UUID  # noqa: E402

import pytest  # noqa: E402

from quiz_management.models.question import (  # noqa: E402
    OptionCreate,
    OptionUpdate,
    QuestionCreate,
    QuestionUpdate,
)
from quiz_management.models.quiz import QuizCreate, QuizUpdate  # noqa: E402
from quiz_management.services.question import QuestionService  # noqa: E402
from quiz_management.services.quiz import QuizService  # noqa: E402
from quiz_management.services.session import SessionService  # noqa: E402


@pytest.fixture
def run_async() -> Callable[[Any], Any]:
    def _run(coroutine: Any) -> Any:
        import asyncio

        return asyncio.run(coroutine)

    return _run


@pytest.fixture
def mock_db() -> AsyncMock:
    return AsyncMock()


@pytest.fixture
def quiz_service(mock_db: AsyncMock) -> QuizService:
    return QuizService(mock_db)


@pytest.fixture
def question_service(mock_db: AsyncMock) -> QuestionService:
    return QuestionService(mock_db)


@pytest.fixture
def mock_session_client() -> AsyncMock:
    return AsyncMock()


@pytest.fixture
def session_service(mock_db: AsyncMock, mock_session_client: AsyncMock) -> SessionService:
    return SessionService(mock_db, mock_session_client)


@pytest.fixture
def quiz_create_factory() -> Callable[..., QuizCreate]:
    def _factory(*, title: str = "Quiz title", description: str = "Quiz description") -> QuizCreate:
        return QuizCreate(title=title, description=description)

    return _factory


@pytest.fixture
def quiz_update_factory() -> Callable[..., QuizUpdate]:
    def _factory(*, title: str | None = None, description: str | None = None) -> QuizUpdate:
        return QuizUpdate(title=title, description=description)

    return _factory


@pytest.fixture
def option_create_factory() -> Callable[..., OptionCreate]:
    def _factory(
        *,
        text: str = "Option",
        order_index: int = 0,
        is_correct: bool = False,
    ) -> OptionCreate:
        return OptionCreate(text=text, order_index=order_index, is_correct=is_correct)

    return _factory


@pytest.fixture
def option_update_factory() -> Callable[..., OptionUpdate]:
    def _factory(
        *,
        option_id: UUID | None = None,
        text: str | None = None,
        order_index: int | None = None,
        is_correct: bool | None = None,
    ) -> OptionUpdate:
        return OptionUpdate(
            id=option_id,
            text=text,
            order_index=order_index,
            is_correct=is_correct,
        )

    return _factory


@pytest.fixture
def question_create_factory(
    option_create_factory: Callable[..., OptionCreate],
) -> Callable[..., QuestionCreate]:
    def _factory(
        *,
        text: str = "Question text",
        selection_type: str = "single",
        time_limit_seconds: int = 15,
        order_index: int = 0,
        options: Any = None,
    ) -> QuestionCreate:
        return QuestionCreate(
            text=text,
            selection_type=selection_type,
            time_limit_seconds=time_limit_seconds,
            order_index=order_index,
            options=options
            or [
                option_create_factory(text="A", order_index=0, is_correct=True),
                option_create_factory(text="B", order_index=1, is_correct=False),
                option_create_factory(text="C", order_index=2, is_correct=False),
                option_create_factory(text="D", order_index=3, is_correct=False),
            ],
        )

    return _factory


@pytest.fixture
def question_update_factory(
    option_update_factory: Callable[..., OptionUpdate],
) -> Callable[..., QuestionUpdate]:
    def _factory(
        *,
        text: str | None = None,
        selection_type: str | None = None,
        time_limit_seconds: int | None = None,
        order_index: int | None = None,
        options: Any = None,
    ) -> QuestionUpdate:
        return QuestionUpdate(
            text=text,
            selection_type=selection_type,
            time_limit_seconds=time_limit_seconds,
            order_index=order_index,
            options=options,
        )

    return _factory
