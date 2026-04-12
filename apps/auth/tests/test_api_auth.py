from datetime import UTC, datetime, timedelta
from unittest.mock import AsyncMock, patch
from uuid import uuid7

import pytest
from fastapi import HTTPException, Request, Response, status

from quiz_auth.api.auth import get_current_profile, login, logout, refresh_token, register
from quiz_auth.models.token import AccessToken, LoginResponse
from quiz_auth.models.users import User, UserLogin, UserPublic


class MockTokenPair:
    def __init__(self):
        now_utc = datetime.now(UTC)
        self.access_token = "access_token-123"  # noqa: S105
        self.refresh_token = "refresh_token-456"  # noqa: S105
        self.token_type = "Bearer"  # noqa: S105
        self.access_expires_in = 900
        self.refresh_expires_in = 604800
        self.access_expires_at = now_utc + timedelta(minutes=15)
        self.refresh_expires_at = now_utc + timedelta(weeks=1)
        self.session_id = uuid7()


@pytest.fixture
def mock_auth_service():
    with patch("quiz_auth.api.auth.AuthService") as mock_auth_service:
        service = AsyncMock()
        mock_auth_service.return_value = service
        yield service


@pytest.fixture
def fake_user() -> User:
    return User(
        id=uuid7(),
        nickname="testuser",
        email="test@example.com",
        role="host",
        token_version=0,
        created_at=datetime.now(UTC),
        updated_at=datetime.now(UTC),
        last_login_at=None,
        deleted_at=None,
        password_hash="hashed_password",  # noqa: S106
    )


@pytest.fixture
def fake_user_public(fake_user) -> UserPublic:
    return UserPublic(
        id=fake_user.id,
        nickname=fake_user.nickname,
        email=fake_user.email,
        role=fake_user.role,
    )


@pytest.fixture
def token_pair():
    return MockTokenPair()


@pytest.fixture
def user_login_factory():
    def _factory(
        *,
        email: str = "user@example.com",
        password: str = "secret_password",  # noqa: S107
    ) -> UserLogin:
        return UserLogin(email=email, password=password)

    return _factory


def test_register_success(mock_auth_service, user_create_factory, fake_user_public, run_async):
    """Test successful user registration"""
    user_data = user_create_factory()
    mock_auth_service.registry_user = AsyncMock(return_value=fake_user_public)

    result = run_async(register(user_data, AsyncMock()))

    assert result == fake_user_public
    mock_auth_service.registry_user.assert_called_once_with(user_data)


def test_duplicate_email(mock_auth_service, user_create_factory, run_async):
    """Test registration with existing email raises conflict"""
    user_data = user_create_factory()
    mock_auth_service.registry_user = AsyncMock(
        side_effect=HTTPException(
            status_code=status.HTTP_409_CONFLICT, detail="User with this email already exists"
        )
    )

    with pytest.raises(HTTPException) as e:
        run_async(register(user_data, AsyncMock()))

    assert e.value.status_code == 409
    assert "email already exists" in e.value.detail.lower()


def test_login_success(
    mock_auth_service, user_login_factory, token_pair, fake_user_public, run_async
):
    """Test successful user login sets refresh cookie"""
    user_data = user_login_factory()
    mock_auth_service.login_user = AsyncMock(return_value=(token_pair, fake_user_public))

    response = Response()
    result = run_async(login(user_data, response, AsyncMock()))

    assert isinstance(result, LoginResponse)
    assert result.access_token == token_pair.access_token
    assert result.user == fake_user_public


def test_login_invalid_credentials(mock_auth_service, user_login_factory, run_async):
    """Test login with invalid credentials return 401"""
    user_data = user_login_factory()
    mock_auth_service.login_user = AsyncMock(
        side_effect=HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Incorrect email or password"
        )
    )

    with pytest.raises(HTTPException) as e:
        run_async(login(user_data, Response(), AsyncMock()))

    assert e.value.status_code == 401


def test_login_user_deactivated(mock_auth_service, user_login_factory, run_async):
    """Test login with deactivated user returns 403"""
    user_data = user_login_factory()
    mock_auth_service.login_user = AsyncMock(
        side_effect=HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="User is deactivated"
        )
    )

    with pytest.raises(HTTPException) as e:
        run_async(login(user_data, Response(), AsyncMock()))

    assert e.value.status_code == 403
    assert e.value.detail == "User is deactivated"


def _create_request_with_cookies(cookies: dict) -> Request:
    """Helper to create Request with cookies"""
    scope = {
        "type": "http",
        "method": "POST",
        "headers": [(b"cookie", b"; ".join([f"{k}={v}".encode() for k, v in cookies.items()]))],
    }
    request = Request(scope)
    return request


def test_refresh_token_success(mock_auth_service, token_pair, run_async):
    """Test successful token refresh with valid cookies"""
    mock_auth_service.refresh_tokens = AsyncMock(return_value=token_pair)

    request = _create_request_with_cookies({"refresh_token": "valid_token"})
    response = Response()
    result = run_async(refresh_token(request, response, AsyncMock()))

    assert isinstance(result, AccessToken)
    assert result.access_token == token_pair.access_token
    mock_auth_service.refresh_tokens.assert_called_once_with("valid_token")


def test_refresh_token_missing_cookie(run_async):
    """Test refresh token without cookies returns 401"""

    request = _create_request_with_cookies({})

    with pytest.raises(HTTPException) as e:
        run_async(refresh_token(request, Response(), AsyncMock()))

    assert e.value.status_code == 401
    assert e.value.detail == "Missing refresh token"


def test_refresh_token_invalid(mock_auth_service, run_async):
    """Test refresh with invalid token returns 401"""

    mock_auth_service.refresh_tokens = AsyncMock(
        side_effect=HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid refresh token"
        )
    )

    request = _create_request_with_cookies({"refresh_token": "invalid_token"})

    with pytest.raises(HTTPException) as e:
        run_async(refresh_token(request, Response(), AsyncMock()))

    assert e.value.status_code == 401


def test_refresh_token_expired(mock_auth_service, run_async):
    """Test refresh with expired token returns 401"""

    mock_auth_service.refresh_tokens = AsyncMock(
        side_effect=HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Refresh token expired"
        )
    )

    request = _create_request_with_cookies({"refresh_token": "expired_token"})

    with pytest.raises(HTTPException) as e:
        run_async(refresh_token(request, Response(), AsyncMock()))

    assert e.value.status_code == 401


def test_get_current_pofile_success(fake_user, run_async):
    """Test getting current user pofile"""

    result = run_async(get_current_profile(fake_user))

    assert result.id == fake_user.id
    assert result.nickname == fake_user.nickname
    assert result.email == fake_user.email
    assert result.role == fake_user.role


def test_logout_success(mock_auth_service, fake_user, run_async):
    """Test successful logout clear cookie"""
    mock_auth_service.logout_user = AsyncMock(return_value=None)

    response = Response()

    result = run_async(logout(response, fake_user, AsyncMock()))

    assert result == {"status": "ok"}
    mock_auth_service.logout_user.assert_called_once_with(fake_user)
