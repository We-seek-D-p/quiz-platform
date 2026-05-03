from typing import Annotated
from uuid import UUID

from fastapi import Depends, Header, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from starlette import status

from quiz_management.core.database import get_session
from quiz_management.models.question import Question
from quiz_management.models.quiz import Quiz
from quiz_management.repositories.question_repository import QuestionRepository
from quiz_management.repositories.quiz_repository import QuizRepository
from quiz_management.services.question import QuestionService
from quiz_management.services.quiz import QuizService
from quiz_management.services.session import SessionService
from quiz_management.services.session_client import SessionServiceClient


async def get_current_user_id(
    user_id: Annotated[UUID | None, Header(alias="X-User-Id")] = None,
) -> UUID:
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


async def get_valid_question(
    question_id: UUID,
    quiz: Annotated[Quiz, Depends(get_valid_quiz)],
    db: Annotated[AsyncSession, Depends(get_session)],
) -> Question:
    repo = QuestionRepository(db)
    question = await repo.get_by_id(question_id)
    if not question or question.quiz_id != quiz.id:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Question not found")
    return question


async def get_question_service(
    db: Annotated[AsyncSession, Depends(get_session)],
) -> QuestionService:
    return QuestionService(db)


async def get_session_service(
    db: Annotated[AsyncSession, Depends(get_session)],
) -> SessionService:
    return SessionService(db)


def get_session_client() -> SessionServiceClient:
    return SessionServiceClient()
