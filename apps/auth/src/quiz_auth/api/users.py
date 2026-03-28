from fastapi import APIRouter, Depends, status, Body
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_auth.core.database import get_session
from quiz_auth.core.dependencies import get_current_user
from quiz_auth.models.users import User, UserPublic, UserUpdate
from quiz_auth.services.user_service import UserService

router = APIRouter(prefix="/users", tags=["users"])


@router.patch("/me", response_model=UserPublic)
async def update_me(
    user_data: UserUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_session),
):
    service = UserService(db)
    return await service.update_profile(current_user.id, user_data)


@router.post("/me/change-password", status_code=status.HTTP_204_NO_CONTENT)
async def change_my_password(
    current_password: str = Body(..., embed=True),
    new_password: str = Body(..., embed=True),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_session),
):
    service = UserService(db)
    await service.change_password(current_user.id, current_password, new_password)
    return None


@router.delete("/me", status_code=status.HTTP_204_NO_CONTENT)
async def delete_my_account(
    current_user: User = Depends(get_current_user), db: AsyncSession = Depends(get_session)
):
    user_repo = UserService(db)
    await user_repo.delete_user(current_user, current_user.id)
    return None
