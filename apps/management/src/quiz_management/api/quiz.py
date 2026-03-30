from uuid import UUID

from apps.management.src.quiz_management.core.dependencies import (
    get_current_user_id,
    get_quiz_service,
    get_valid_quiz,
)
from apps.management.src.quiz_management.models.quiz import Quiz, QuizCreate, QuizPublic, QuizUpdate
from apps.management.src.quiz_management.services.quiz import QuizService
from fastapi import APIRouter, Depends, status
from sqlalchemy.sql.annotation import Annotated

router = APIRouter(prefix="/quizzes", tags=["Quizzes"])


@router.post("/", response_model=QuizPublic, status_code=status.HTTP_201_CREATED)
async def create_quiz(
    data: QuizCreate,
    user_id: Annotated[UUID, Depends(get_current_user_id)],
    service: Annotated[QuizService, Depends(get_quiz_service)],
):
    return await service.create_quiz(user_id, data)


@router.get("/{quiz_id}", response_model=QuizPublic)
async def get_quiz(quiz: Annotated[Quiz, Depends(get_valid_quiz)]):
    return quiz


@router.patch("/{quiz_id}", response_model=QuizPublic)
async def update_quiz(
    data: QuizUpdate,
    quiz: Annotated[Quiz, Depends(get_valid_quiz)],
    service: Annotated[QuizService, Depends(get_quiz_service)],
):
    return await service.update_quiz(quiz, data)


@router.delete("/{quiz_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_quiz(
    quiz: Annotated[Quiz, Depends(get_valid_quiz)],
    service: Annotated[QuizService, Depends(get_quiz_service)],
):
    await service.delete_quiz(quiz)
