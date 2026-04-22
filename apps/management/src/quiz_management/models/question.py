from datetime import UTC, datetime
from uuid import UUID, uuid7

from pydantic import BaseModel, field_validator
from sqlmodel import Field, Relationship, SQLModel

from quiz_management.models.quiz import Quiz


def get_utc_now():
    return datetime.now(UTC).replace(tzinfo=None)


def validate_options_list(v: list | None):
    if v is None:
        return v
    if not any(opt.is_correct for opt in v):
        raise ValueError("At least one option must be correct")
    return v


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


class OptionPublic(BaseModel):
    id: UUID
    text: str
    order_index: int
    is_correct: bool


class OptionCreate(BaseModel):
    text: str
    order_index: int
    is_correct: bool = False


class QuestionCreate(BaseModel):
    text: str
    selection_type: str = "single"
    time_limit_seconds: int = 15
    order_index: int
    options: list[OptionCreate] = Field(..., min_length=4, max_length=4)

    @field_validator("options")
    @classmethod
    def validate_opts(cls, v):
        return validate_options_list(v)


class OptionUpdate(BaseModel):
    id: UUID | None = None
    text: str | None = None
    order_index: int | None = None
    is_correct: bool | None = None


class QuestionUpdate(BaseModel):
    text: str | None = None
    selection_type: str | None = None
    time_limit_seconds: int | None = None
    order_index: int | None = None
    options: list[OptionUpdate] | None = Field(None, min_length=4, max_length=4)

    @field_validator("options")
    @classmethod
    def validate_opts(cls, v):
        return validate_options_list(v)


class QuestionPublic(BaseModel):
    id: UUID
    quiz_id: UUID
    text: str
    selection_type: str
    time_limit_seconds: int
    order_index: int
    options: list[OptionPublic]
