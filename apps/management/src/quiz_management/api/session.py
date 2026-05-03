from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, Header, status
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.core.database import get_session
from quiz_management.core.dependencies import (
    get_current_user_id,
    get_session_service,
    get_valid_quiz,
)
from quiz_management.models.session import SessionCreate, SessionPublic
from quiz_management.service.session import SessionService

router = APIRouter(prefix="/sessions", tags=["Sessions"])


@router.post("/", response_model=SessionPublic, status_code=status.HTTP_201_CREATED)
async def create_session(
    data: SessionCreate,
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    service: Annotated[SessionService, Depends(get_session_service)],
    x_idempotency_key: Annotated[str, Header(alias="Idempotency-Key")],
    db: Annotated[AsyncSession, Depends(get_session)],
):
    quiz = await get_valid_quiz(quiz_id=data.quiz_id, user_id=user_id, db=db)

    return await service.create_session(
        quiz=quiz, user_id=user_id, idempotency_key=x_idempotency_key
    )
