from collections.abc import Sequence
from uuid import UUID

from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.models.quiz import Quiz


class QuizRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def get_by_id(self, quiz_id: UUID) -> Quiz | None:
        statement = select(Quiz).where(Quiz.id == quiz_id, Quiz.deleted_at == None)  # noqa: E711
        result = await self.db.exec(statement)
        return result.first()

    async def get_by_owner_id(self, owner_id: UUID) -> Sequence[Quiz]:
        statement = select(Quiz).where(Quiz.owner_id == owner_id).order_by(Quiz.created_at.desc())
        result = await self.db.exec(statement)
        return result.all()

    async def save(self, quiz: Quiz) -> Quiz:
        self.db.add(quiz)
        await self.db.commit()
        await self.db.refresh(quiz)
        return quiz
