from fastapi import APIRouter, Depends, Body
from sqlmodel.ext.asyncio.session import AsyncSession
from backend.core.database import get_session
from backend.core.dependencies import get_current_user
from backend.models.token import Token
from backend.models.users import UserPublic, UserCreate, UserLogin, User, UserUpdate
from backend.services.auth_service import AuthService
from backend.services.user_service import UserService

router = APIRouter(prefix="/users", tags=["users"])


@router.post("/register", response_model=UserPublic)
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
async def refresh_token(request_refresh_token: str = Body(...), db: AsyncSession = Depends(get_session)):
    auth_service = AuthService(db)
    new_tokens = await auth_service.refresh_tokens(request_refresh_token)
    return new_tokens


@router.get("/{user_id}", response_model=UserPublic)
async def get_user(
        user_id: int,
        current_user: User = Depends(get_current_user),
        db: AsyncSession = Depends(get_session)
):
    service = UserService(db)
    user = await service.get_user_by_id(user_id, current_user)
    return user


@router.patch("/{user_id}", response_model=UserPublic)
async def update_user(
        user_id: int,
        user_update: UserUpdate,
        current_user: User = Depends(get_current_user),
        db: AsyncSession = Depends(get_session)
):
    service = UserService(db)
    update_data = user_update.model_dump(exclude_unset=True)
    updated = await service.update_user(update_data, user_id, current_user)
    return updated


@router.delete("/{user_id}", status_code=204)
async def delete_user(
        user_id: int,
        current_user: User = Depends(get_current_user),
        db: AsyncSession = Depends(get_session)
):
    service = UserService(db)
    await service.delete_user(current_user, user_id)
    return None
