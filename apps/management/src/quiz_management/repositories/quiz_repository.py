import datetime
from uuid import UUID
from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession
from quiz_management.models.quiz import Quiz

QUIZ_TABLE = Quiz.__table__
_DELETED_AT_COLUMN = QUIZ_TABLE.c.deleted_at


class QuizzRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    # TODO - maybe we need to remove self.db.refresh?

    async def get_by_id(self, quizz_id: UUID) -> Quiz | None:
        query = select(Quiz).where(Quiz.id == quizz_id, _DELETED_AT_COLUMN.is_(None))
        if not query:
            return None
        result = await self.db.exec(query)
        return result.first()

    async def create_quiz(self, quiz: Quiz) -> Quiz:
        self.db.add(quiz)
        await self.db.commit()
        await self.db.refresh(quiz)
        return quiz

    async def update_quiz(self, quiz_id: UUID, quiz_data: dict) -> Quiz | None:
        quiz = await self.get_by_id(quiz_id)
        if not quiz:
            return None
        for key, value in quiz_data.items():
            setattr(quiz, key, value)
        self.db.add(quiz)
        await self.db.commit()
        await self.db.refresh(quiz)
        return quiz

    async def delete_quiz(self, quiz_id: UUID) -> bool:
        quiz = await self.get_by_id(quiz_id)
        if not quiz:
            return False
        quiz.deleted_at = datetime.datetime.now(datetime.UTC)
        self.db.add(quiz)
        await self.db.commit()
        return True
