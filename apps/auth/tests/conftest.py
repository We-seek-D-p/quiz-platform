from collections.abc import Callable
from dataclasses import dataclass
from typing import Any
from unittest.mock import AsyncMock
from uuid import uuid7

import pytest
from quiz_auth.models.users import UserCreate, UserUpdate
from quiz_auth.repositories.user_repository import UserRepository
from quiz_auth.services.user_service import UserService


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
def user_repository(mock_db: AsyncMock) -> UserRepository:
    return UserRepository(mock_db)


@pytest.fixture
def user_service(mock_db: AsyncMock) -> UserService:
    return UserService(mock_db)


@dataclass(frozen=True)
class FakeUser:
    id: object


@pytest.fixture
def user_create_factory() -> Callable[..., UserCreate]:
    def _factory(
        *,
        nickname: str = "test-user",
        email: str = "user@example.com",
        password: str = "secret-password",  # noqa: S107
    ) -> UserCreate:
        return UserCreate(nickname=nickname, email=email, password=password)

    return _factory


@pytest.fixture
def user_update_factory() -> Callable[..., UserUpdate]:
    def _factory(
        *,
        nickname: str | None = None,
        email: str | None = None,
        password: str | None = None,
        role: str | None = None,
    ) -> UserUpdate:
        return UserUpdate(nickname=nickname, email=email, password=password, role=role)

    return _factory


@pytest.fixture
def fake_user_factory() -> Callable[..., FakeUser]:
    def _factory(*, user_id: object | None = None) -> FakeUser:
        return FakeUser(id=user_id or uuid7())

    return _factory


@pytest.fixture
def token_factory() -> Callable[..., str]:
    def _factory(*, value: str | None = None) -> str:
        return value or "test-refresh-token"

    return _factory
