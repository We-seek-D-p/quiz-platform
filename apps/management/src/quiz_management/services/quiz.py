from datetime import UTC, datetime
from uuid import UUID

from sqlalchemy.ext.asyncio.session import AsyncSession

from quiz_management.models.quiz import Quiz, QuizCreate, QuizUpdate
from quiz_management.repositories.quiz_repository import QuizRepository


class QuizService:
    def __init__(self, db: AsyncSession):
        self.repository = QuizRepository(db)

    async def create_quiz(self, user_id: UUID, data: QuizCreate) -> Quiz:
        quiz = Quiz(**data.model_dump(), owner_id=user_id)
        return await self.repository.save(quiz)

    async def get_quizzes(self, user_id: UUID) -> list[Quiz]:
        return await self.repository.get_by_owner_id(user_id)

    async def update_quiz(self, quiz: Quiz, data: QuizUpdate) -> Quiz:
        update_dict = data.model_dump(exclude_unset=True)

        for key, value in update_dict.items():
            setattr(quiz, key, value)

        if update_dict:
            quiz.updated_at = datetime.now(UTC)

        return await self.repository.save(quiz)

    async def delete_quiz(self, quiz: Quiz) -> None:
        quiz.deleted_at = datetime.now(UTC)
        await self.repository.save(quiz)
