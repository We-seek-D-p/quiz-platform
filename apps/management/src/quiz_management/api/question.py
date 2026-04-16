from typing import Annotated, Any

from fastapi import APIRouter, Depends, status

from quiz_management.core.dependencies import (
    get_question_service,
    get_valid_question,
    get_valid_quiz,
)
from quiz_management.models.question import Question, QuestionCreate, QuestionPublic, QuestionUpdate
from quiz_management.models.quiz import Quiz
from quiz_management.services.question import QuestionService

router = APIRouter(prefix="/quizzes/{quiz_id}/questions", tags=["Questions"])


@router.post("/", response_model=QuestionPublic, status_code=status.HTTP_201_CREATED)
async def create_question(
    quiz: Annotated[Any, Depends(get_valid_quiz)],
    data: QuestionCreate,
    service: Annotated[QuestionService, Depends(get_question_service)],
):
    return await service.create_question(data, quiz.id)


@router.get("/{question_id}", response_model=QuestionPublic, status_code=status.HTTP_200_OK)
async def get_question(question: Annotated[Question, Depends(get_valid_question)]):
    return question


@router.get("/", response_model=list[QuestionPublic])
async def get_questions(
    quiz: Annotated[Quiz, Depends(get_valid_quiz)],
    service: Annotated[QuestionService, Depends(get_question_service)],
):
    return await service.get_quiz_questions(quiz.id)


@router.patch("/{question_id}", response_model=QuestionPublic)
async def update_question(
    data: QuestionUpdate,
    question: Annotated[Question, Depends(get_valid_question)],
    service: Annotated[QuestionService, Depends(get_question_service)],
):
    return await service.update_question(question, data)


@router.delete("/{question_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_question(
    question: Annotated[Question, Depends(get_valid_question)],
    service: Annotated[QuestionService, Depends(get_question_service)],
):
    await service.delete_question(question)
