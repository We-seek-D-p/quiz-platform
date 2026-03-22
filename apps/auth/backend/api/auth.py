from fastapi import APIRouter, Body, Depends, Response, status
from sqlmodel.ext.asyncio.session import AsyncSession

from backend.core.database import get_session
from backend.core.dependencies import get_current_user
from backend.models.token import Token
from backend.models.users import User, UserCreate, UserLogin, UserPublic
from backend.services.auth_service import AuthService


router = APIRouter(prefix="/auth", tags=["auth"])


@router.post("/register", response_model=UserPublic, status_code=status.HTTP_201_CREATED)
async def register(user_data: UserCreate, db: AsyncSession = Depends(get_session)):
    auth_service = AuthService(db)
    user = await auth_service.registry_user(user_data)
    return user


@router.post("/login", response_model=Token)
async def login(user_data: UserLogin, db: AsyncSession = Depends(get_session)):
    auth_service = AuthService(db)
    tokens = await auth_service.login_user(user_data)
    return tokens


@router.post("/refresh", response_model=Token)
async def refresh_token(
    refresh_token: str = Body(..., embed=True, alias="refresh_token"),
    db: AsyncSession = Depends(get_session),
):
    auth_service = AuthService(db)
    new_tokens = await auth_service.refresh_tokens(refresh_token)
    return new_tokens


@router.get("/me", response_model=UserPublic)
async def get_current_profile(current_user: User = Depends(get_current_user)):
    return current_user


@router.post("/logout")
async def logout(
    response: Response,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_session),
):
    auth_service = AuthService(db)
    await auth_service.logout_user(current_user)
    response.delete_cookie(
        key="refresh_token",
        path="/auth/refresh",
        httponly=True,
        secure=True,
        samesite="strict",
    )
    return {"status": "ok"}
