from uuid import UUID, uuid7
from sqlmodel import Field, SQLModel, Relationship

from models.quizz import Quiz


class QuestionBase(SQLModel):
    quiz_id: UUID = Field(foreign_key="quizzes.id", index=True)
    text: str
    selection_type: str = Field(default="single", max_length=10)
    time_limit_seconds: int = Field(default=15)
    order_index: int

class Question(QuestionBase, table=True):
    __tablename__ = "questions"
    id: UUID = Field(default_factory=uuid7)

    quiz: Quiz = Relationship(back_populates="questions")
    options: list[QuestionOption] = Relationship(back_populates="question")


class QuestionOption(SQLModel, table=True):
    __tablename__ = "question_options"
    id: UUID = Field(default_factory=uuid7, primary_key=True)
    question_id: UUID = Field(foreign_key="questions.id")
    text: str
    order_index: int

    question: Question = Relationship(back_populates="options")

class QuestionCorrectOption(SQLModel, table=True):
    __tablename__ = "question_correct_options"
    question_id: UUID = Field(foreign_key="questions.id", primary_key=True)
    option_id: UUID = Field(foreign_key="question_option.id", primary_key=True)
