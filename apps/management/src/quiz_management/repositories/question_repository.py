from uuid import UUID

from models.question import QuestionOption
from sqlalchemy.orm import selectinload
from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.models.question import Question


class QuestionRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def get_by_id(self, question_id: UUID) -> Question | None:
        statement = (
            select(Question)
            .where(
                Question.id == question_id,
                Question.deleted_at == None,  # noqa: E711
            )
            .options(selectinload(Question.options).where(QuestionOption.deleted_at == None))  # noqa: E711
        )
        result = await self.db.exec(statement)
        return result.first()

    async def save(self, question: Question) -> Question:
        self.db.add(question)
        await self.db.commit()
        await self.db.refresh(question)
        return question
