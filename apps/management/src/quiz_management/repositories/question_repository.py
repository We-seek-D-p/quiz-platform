import datetime
from uuid import UUID

from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.models.question import Question
from quiz_management.models.quiz import Quiz

QUESTION_TABLE = Question.__table__
_DELETED_AT_COLUMN = QUESTION_TABLE.c.deleted_at


class QuestionRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def get_by_id(self, question_id: UUID) -> Question | None:
        query = select(self.db).where(Quiz.id == question_id, _DELETED_AT_COLUMN.is_(None))
        if not query:
            return None
        result = await self.db.exec(query)
        return result.first()

    async def create_question(self, question: Question) -> Question | None:
        self.db.add(question)
        await self.db.commit()
        await self.db.refresh(question)
        return question

    async def update_question(self, question_id: UUID, question_data: dict):
        question = self.get_by_id(question_id)
        if not question:
            return None
        for key, value in question_data.items():
            setattr(question, key, value)
        self.db.add(question)
        await self.db.commit()
        await self.db.refresh(question)
        return question

    async def delete_question(self, question_id: UUID) -> bool:
        question = self.get_by_id(question_id)
        if not question:
            return False
        question.deleted_at = datetime.datetime.now(datetime.UTC)
        self.db.add(question)
        await self.db.commit()
        return True
