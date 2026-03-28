from datetime import datetime, timezone
from uuid import UUID, uuid7
from sqlmodel import Field, SQLModel, Relationship
from question import Question
from session import GameSession

def get_utc_now():
    return datetime.now(timezone.utc)

class QuizBase(SQLModel):
    title: str = Field(max_length=256)
    description: str = Field(max_length=512)
    owner_id: UUID

class Quiz(QuizBase, table=True):
    __tablename__ = "quizzes"
    id: UUID = Field(default_factory=uuid7, primary_key=True)
    created_at: datetime = Field(default_factory=get_utc_now)
    updated_at: datetime = Field(default_factory=get_utc_now)
    deleted_at: datetime | None  = Field(default = None)

    questions: list[Question] = Relationship(back_populates="quiz")
    sessions: list[GameSession] = Relationship(back_populates="quiz")
