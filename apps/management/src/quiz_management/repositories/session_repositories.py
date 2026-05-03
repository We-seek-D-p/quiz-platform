from typing import TypeVar
from uuid import UUID

from sqlalchemy.orm import selectinload
from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_management.models.question import Question
from quiz_management.models.quiz import Quiz
from quiz_management.models.session import GameSession, SessionParticipant

TSession = TypeVar("TSession", GameSession, SessionParticipant)


class SessionRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def save_session(self, session: TSession) -> None:
        self.db.add(session)
        await self.db.commit()
        await self.db.refresh(session)

    async def get_session_with_quiz(self, session_id: UUID) -> GameSession | None:
        statement = (
            select(GameSession)
            .where(GameSession.id == session_id)
            .options(
                selectinload(GameSession.quiz)
                .selectinload(Quiz.questions)
                .selectinload(Question.options)
            )
        )
        result = await self.db.exec(statement)
        return result.first()
