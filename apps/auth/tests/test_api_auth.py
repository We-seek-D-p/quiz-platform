from unittest.mock import AsyncMock, patch
from uuid import uuid7
from datetime import datetime, timedelta, timezone

import pytest
from fastapi import HTTPException, Response, Request, status

from quiz_auth.models.token import AccessToken, LoginResponse
from quiz_auth.models.users import User, UserPublic, UserLogin


class MockTokenPair:
    def __init__(self):
        now_utc = datetime.now(timezone.utc)
        self.access_token = 'access_token-123'
        self.refresh_token = 'refresh_token-456'
        self.token_type = 'Bearer'
        self.access_expires_in = 900
        self.refresh_expires_in = 604800
        self.access_expires_at = now_utc + timedelta(minutes=15)
        self.refresh_expires_at = now_utc + timedelta(weeks=1)
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
def fake_user_public(fake_user) -> UserPublic:
    return UserPublic(
        id = fake_user.id,
        nickname = fake_user.nickname,
        email = fake_user.email,
        role = fake_user.role,
    )


@pytest.fixture
def token_pair():
    return MockTokenPair()


@pytest.fixture
def user_login_factory():
    def _factory(*, email: str = 'user@example.com', password: str = 'secret_password') -> UserLogin:
        return UserLogin(email=email, password=password)
    return _factory


def test_register_success(mock_auth_service, user_create_factory, fake_user_public, run_async):
    """Test successful user registration"""
    user_data = user_create_factory()
    mock_auth_service.registry_user = AsyncMock(return_value=fake_user_public)

    from quiz_auth.api.auth import register
    result = run_async(register(user_data, AsyncMock()))

    assert result == fake_user_public
    mock_auth_service.registry_user.assert_called_once_with(user_data)


def test_duplicate_email(mock_auth_service, user_create_factory, run_async):
    """Test registration with existing email raises conflict"""
    user_data = user_create_factory()
    mock_auth_service.registry_user = AsyncMock(
        side_effect=HTTPException(status_code=status.HTTP_409_CONFLICT, detail='User with this email already exists')
    )

    from quiz_auth.api.auth import register
    with pytest.raises(HTTPException) as e:
        run_async(register(user_data, AsyncMock()))

    assert e.value.status_code == 409
    assert 'email already exists' in e.value.detail.lower()


def test_login_success(mock_auth_service, user_login_factory, token_pair, fake_user_public, run_async):
    """Test successful user login sets refresh cookie"""
    user_data = user_login_factory()
    mock_auth_service.login_user = AsyncMock(return_value=(token_pair, fake_user_public))

    from quiz_auth.api.auth import login
    response = Response()
    result = run_async(login(user_data, response, AsyncMock()))

    assert isinstance(result, LoginResponse)
    assert result.access_token == token_pair.access_token
    assert result.user == fake_user_public


def test_login_invalid_credentials(mock_auth_service, user_login_factory, run_async):
    """Test login with invalid credentials return 401"""
    user_data = user_login_factory()
    mock_auth_service.login_user = AsyncMock(
        side_effect=HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail='Incorrect email or password')
    )

    from quiz_auth.api.auth import login
    with pytest.raises(HTTPException) as e:
        run_async(login(user_data, Response(), AsyncMock()))

    assert e.value.status_code == 401


def test_login_user_deactivated(mock_auth_service, user_login_factory, run_async):
    """Test login with deactivated user returns 403"""
    user_data = user_login_factory()
    mock_auth_service.login_user = AsyncMock(
        side_effect=HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail='User is deactivated')
    )

    from quiz_auth.api.auth import login
    with pytest.raises(HTTPException) as e:
        run_async(login(user_data, Response(), AsyncMock()))

    assert e.value.status_code == 403
    assert e.value.detail == 'User is deactivated'


def _create_request_with_cookies(cookies: dict) -> Request:
    """Helper to create Request with cookies"""
    scope = {
        'type': 'http',
        'method': 'POST',
        'headers': [(b"cookie", b"; ".join([f'{k}={v}'.encode() for k, v in cookies.items()]))]
    }
    request = Request(scope)
    return request


def test_refresh_token_success(mock_auth_service, token_pair, run_async):
    """Test successful token refresh with valid cookies"""
    mock_auth_service.refresh_tokens = AsyncMock(return_value=token_pair)

    from quiz_auth.api.auth import refresh_token

    request = _create_request_with_cookies({'refresh_token': 'valid_token'})
    response = Response()
    result = run_async(refresh_token(request, response, AsyncMock()))

    assert isinstance(result, AccessToken)
    assert result.access_token == token_pair.access_token
    mock_auth_service.refresh_tokens.assert_called_once_with('valid_token')


def test_refresh_token_missing_cookie(run_async):
    """Test refresh token without cookies returns 401"""
    from quiz_auth.api.auth import refresh_token

    request = _create_request_with_cookies({})

    with pytest.raises(HTTPException) as e:
        run_async(refresh_token(request, Response(), AsyncMock()))

    assert e.value.status_code == 401
    assert e.value.detail == 'Missing refresh token'
