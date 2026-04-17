from datetime import UTC, datetime
from uuid import UUID, uuid7

from sqlmodel import Field, Relationship, SQLModel


def get_utc_now():
    return datetime.now(UTC).replace(tzinfo=None)


class QuizBase(SQLModel):
    title: str = Field(max_length=256)
    description: str = Field(max_length=512)


class Quiz(QuizBase, table=True):
    __tablename__ = "quizzes"
    __table_args__ = {"schema": "management"}
    id: UUID = Field(default_factory=uuid7, primary_key=True)
    owner_id: UUID
    created_at: datetime = Field(default_factory=get_utc_now)
    updated_at: datetime = Field(
        default_factory=get_utc_now, sa_column_kwargs={"onupdate": get_utc_now}
    )
    deleted_at: datetime | None = Field(default=None)

    questions: list[Question] = Relationship(back_populates="quiz")  # noqa: F821
    sessions: list[GameSession] = Relationship(back_populates="quiz")  # noqa: F821


class QuizCreate(QuizBase):
    pass


class QuizUpdate(SQLModel):
    title: str | None = None
    description: str | None = None


class QuizPublic(QuizBase):
    id: UUID
    created_at: datetime
    updated_at: datetime
