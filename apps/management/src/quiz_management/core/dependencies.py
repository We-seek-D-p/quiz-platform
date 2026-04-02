from uuid import UUID

from core.database import get_session
from fastapi import Depends, Header, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.sql.annotation import Annotated
from starlette import status

from quiz_management.models.quiz import Quiz
from quiz_management.repositories.quiz_repository import QuizRepository
from quiz_management.services.quiz import QuizService


async def get_current_user_id(user_id: str = Header(None, alias="X-User-Id")) -> str:
    if not user_id:
        raise HTTPException(status_code=401, detail="X-User-Id header missing")
    return user_id


async def get_valid_quiz(
    quiz_id: UUID,
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    db: Annotated[AsyncSession, Depends(get_session)],
) -> Quiz:
    repo = QuizRepository(db)
    quiz = await repo.get_by_id(quiz_id)
    if not quiz or quiz.owner_id != user_id:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Quiz not found")
    return quiz


async def get_quiz_service(db: Annotated[AsyncSession, Depends(get_session)]) -> QuizService:
    return QuizService(db)
