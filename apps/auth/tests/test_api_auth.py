from unittest.mock import AsyncMock, patch
from uuid import uuid7
from datetime import datetime, timedelta, timezone

import pytest


from quiz_auth.models.users import User, UserLogin


class MockTockenPair:
    def __init__(self):
        now_utc = datetime.now(timezone.utc)
        self.access_token = 'access_token-123'
        self.refresh_token = 'refresh_token-456'
        self.token_type = 'Bearer'
        self.access_expires_in = 900
        self.refresh_expires_in = 604800
        self.access_token_expires_at = now_utc + timedelta(minutes=15)
        self.refresh_token_expires_at = now_utc + timedelta(weeks=1)
        self.session_id = uuid7()


@pytest.fixture
def mock_auth_service():
    with patch('quiz_auth.api.auth.AuthService') as MockAuthService:
        service = AsyncMock()
        MockAuthService.return_value = service
        yield service


@pytest.fixture
def fake_user() -> User:
    return User(
        id = uuid7(),
        nickname = 'testuser',
        email = 'test@example.com',
        role = 'host',
        token_version = 0,
        created_at = datetime.now(timezone.utc),
        updated_at = datetime.now(timezone.utc),
        last_login_at = None,
        deleted_at = None,
        password_hash = 'hashed_password',
    )


@pytest.fixture
def fake_user_public() -> User:
    return User(
        id = fake_user.id,
        nickname = fake_user.nickname,
        email = fake_user.email,
        role = fake_user.role,
    )


@pytest.fixture
def token_pair():
    return MockTockenPair()


@pytest.fixture
def user_login_factory():
    def _factory(*, email: str = 'user@example.com', password: str = 'secret_password') -> UserLogin:
        return UserLogin(email=email, password=password)
    return _factory
