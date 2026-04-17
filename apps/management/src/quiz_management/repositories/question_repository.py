from collections.abc import Sequence
from uuid import UUID

from sqlalchemy.orm import selectinload, with_loader_criteria
from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.models.question import Question, QuestionOption


class QuestionRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def get_by_id(self, question_id: UUID) -> Question:
        statement = (
            select(Question)
            .where(Question.id == question_id, Question.deleted_at == None)  # noqa: E711
            .options(
                selectinload(Question.options),
                with_loader_criteria(
                    QuestionOption,
                    QuestionOption.deleted_at == None,  # noqa: E711
                    include_aliases=True,
                ),
            )
        )
        result = await self.db.exec(statement)
        return result.first()

    async def get_by_quiz_id(self, quiz_id: UUID) -> Sequence[Question] | None:
        statement = (
            select(Question)
            .where(Question.quiz_id == quiz_id, Question.deleted_at == None)  # noqa: E711
            .options(
                selectinload(Question.options),
                with_loader_criteria(
                    QuestionOption,
                    QuestionOption.deleted_at == None,  # noqa: E711
                    include_aliases=True,
                ),
            )
            .order_by(Question.order_index)
        )
        result = await self.db.exec(statement)
        return result.unique().all()

    async def save(self, question: Question) -> Question:
        self.db.add(question)
        await self.db.commit()
        return question
