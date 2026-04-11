from datetime import datetime
from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, Request, Response, status
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_auth.core.database import get_session
from quiz_auth.core.dependencies import get_current_user
from quiz_auth.models.token import AccessToken, LoginResponse
from quiz_auth.models.users import User, UserCreate, UserLogin, UserPublic
from quiz_auth.services.auth_service import AuthService

router = APIRouter(prefix="/auth", tags=["auth"])

REFRESH_COOKIE_NAME = "refresh_token"
REFRESH_COOKIE_PATH = "/api/v1/auth/refresh"


def _set_refresh_cookie(response: Response, token: str, max_age: int, expires_at: datetime) -> None:
    response.set_cookie(
        key=REFRESH_COOKIE_NAME,
        value=token,
        max_age=max_age,
        expires=expires_at,
        path=REFRESH_COOKIE_PATH,
        secure=True,
        httponly=True,
        samesite="strict",
    )


def _clear_refresh_cookie(response: Response) -> None:
    response.set_cookie(
        key=REFRESH_COOKIE_NAME,
        value="",
        max_age=0,
        expires=0,
        path=REFRESH_COOKIE_PATH,
        secure=True,
        httponly=True,
        samesite="strict",
    )


@router.post("/register", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def register(user_data: UserCreate, db: Annotated[AsyncSession, Depends(get_session)]):
    auth_service = AuthService(db)
    user = await auth_service.registry_user(user_data)
    return user


@router.post("/login", response_model=LoginResponse)
async def login(
    user_data: UserLogin,
    response: Response,
    db: Annotated[AsyncSession, Depends(get_session)],
):
    auth_service = AuthService(db)
    token_pair, user = await auth_service.login_user(user_data)
    _set_refresh_cookie(
        response,
        token_pair.refresh_token,
        token_pair.refresh_expires_in,
        token_pair.refresh_expires_at,
    )
    return LoginResponse(
        access_token=token_pair.access_token,
        token_type=token_pair.token_type,
        expires_in=token_pair.access_expires_in,
        user=user,
    )


@router.post("/refresh", response_model=AccessToken)
async def refresh_token(
    request: Request,
    response: Response,
    db: Annotated[AsyncSession, Depends(get_session)],
):
    refresh_token_cookie = request.cookies.get(REFRESH_COOKIE_NAME)
    if not refresh_token_cookie:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Missing refresh token"
        )
    auth_service = AuthService(db)
    new_tokens = await auth_service.refresh_tokens(refresh_token_cookie)
    _set_refresh_cookie(
        response,
        new_tokens.refresh_token,
        new_tokens.refresh_expires_in,
        new_tokens.refresh_expires_at,
    )
    return AccessToken(
        access_token=new_tokens.access_token,
        token_type=new_tokens.token_type,
        expires_in=new_tokens.access_expires_in,
    )


@router.get("/me", response_model=UserPublic)
async def get_current_profile(current_user: Annotated[User, Depends(get_current_user)]):
    return current_user


@router.post("/logout")
async def logout(
    response: Response,
    current_user: Annotated[User, Depends(get_current_user)],
    db: Annotated[AsyncSession, Depends(get_session)],
):
    auth_service = AuthService(db)
    await auth_service.logout_user(current_user)
    _clear_refresh_cookie(response)
    return {"status": "ok"}
