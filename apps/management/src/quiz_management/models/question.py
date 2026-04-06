from datetime import UTC, datetime
from uuid import UUID, uuid7

from sqlmodel import Field, Relationship, SQLModel

from quiz_management.models.quiz import Quiz


def get_utc_now():
    return datetime.now(UTC)


class TimestampMixin(SQLModel):
    created_at: datetime = Field(default_factory=get_utc_now)
    updated_at: datetime = Field(
        default_factory=get_utc_now, sa_column_kwargs={"onupdate": get_utc_now}
    )
    deleted_at: datetime | None = Field(default=None)


class QuestionBase(SQLModel):
    quiz_id: UUID = Field(foreign_key="management.quizzes.id", index=True)
    text: str
    selection_type: str = Field(default="single", max_length=10)
    time_limit_seconds: int = Field(default=15)
    order_index: int


class OptionBase(SQLModel):
    text: str
    order_index: int
    is_correct: bool = Field(default=False)


class Question(QuestionBase, TimestampMixin, table=True):
    __tablename__ = "questions"
    __table_args__ = {"schema": "management"}

    id: UUID = Field(default_factory=uuid7, primary_key=True)

    quiz: Quiz = Relationship(back_populates="questions")
    options: list[QuestionOption] = Relationship(back_populates="question")


class QuestionOption(OptionBase, TimestampMixin, table=True):
    __tablename__ = "question_options"
    __table_args__ = {"schema": "management"}

    id: UUID = Field(default_factory=uuid7, primary_key=True)

    question_id: UUID = Field(foreign_key="management.questions.id")
    question: Question = Relationship(back_populates="options")


class OptionPublic(OptionBase):
    id: UUID


class QuestionCreate(QuestionBase):
    options: list[OptionBase]


class OptionUpdate(SQLModel):
    id: UUID | None = None
    text: str | None = None
    order_index: int | None = None
    is_correct: bool | None = None


class QuestionUpdate(SQLModel):
    text: str | None = None
    selection_type: str | None = None
    time_limit_seconds: int | None = None
    order_index: int | None = None
    options: list[OptionUpdate] | None = None


class QuestionPublic(QuestionBase):
    id: UUID
    options: list[OptionPublic]
