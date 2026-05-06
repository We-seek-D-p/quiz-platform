from datetime import UTC, datetime
from enum import StrEnum
from uuid import UUID, uuid7

from pydantic import BaseModel
from sqlalchemy import String
from sqlmodel import Field, Relationship, SQLModel

from quiz_management.models.quiz import Quiz


def get_utc_now():
    return datetime.now(UTC).replace(tzinfo=None)


class SessionStatus(StrEnum):
    INITIALIZING = "initializing"
    LOBBY = "lobby"
    IN_PROGRESS = "in_progress"
    FINISHED = "finished"
    INIT_FAILED = "init_failed"


class GameSession(SQLModel, table=True):
    __tablename__ = "game_sessions"
    __table_args__ = {"schema": "management"}
    id: UUID = Field(default_factory=uuid7, primary_key=True)
    quiz_id: UUID = Field(foreign_key="management.quizzes.id")
    room_code: str | None = Field(default=None, max_length=8)
    host_id: UUID
    status: SessionStatus = Field(
        default=SessionStatus.INITIALIZING,
        sa_type=String,
    )

    created_at: datetime = Field(default_factory=get_utc_now)
    started_at: datetime | None = None
    finished_at: datetime | None = None

    quiz: Quiz = Relationship(back_populates="sessions")
    participants: list[SessionParticipant] = Relationship(back_populates="session")


class SessionParticipant(SQLModel, table=True):
    __tablename__ = "session_participants"
    __table_args__ = {"schema": "management"}
    id: UUID = Field(default_factory=uuid7, primary_key=True)
    session_id: UUID = Field(foreign_key="management.game_sessions.id")
    player_nickname: str = Field(max_length=255)
    score: int = Field(default=0)
    rank: int | None = None

    session: GameSession = Relationship(back_populates="participants")


class SessionCreate(SQLModel):
    quiz_id: UUID


class SessionPublic(SQLModel):
    id: UUID
    quiz_id: UUID
    room_code: str | None
    status: SessionStatus
    host_id: UUID


class SessionBootstrapPublic(SQLModel):
    session_id: UUID
    quiz_id: UUID
    room_code: str | None
    status: SessionStatus
    host_id: UUID


class OptionSnapshotPublic(SQLModel):
    id: UUID
    text: str
    order_index: int
    is_correct: bool


class QuestionSnapshotPublic(SQLModel):
    id: UUID
    text: str
    selection_type: str
    time_limit_seconds: int
    order_index: int
    options: list[OptionSnapshotPublic]


class QuizSnapshotPublic(SQLModel):
    id: UUID
    title: str
    description: str
    questions: list[QuestionSnapshotPublic]


class SessionBootstrap(SQLModel):
    session: SessionBootstrapPublic
    quiz_snapshot: QuizSnapshotPublic


class SessionStatusUpdate(SQLModel):
    status: SessionStatus
    event_id: str
    started_at: datetime | None = None


class ParticipantResult(BaseModel):
    participant_id: UUID
    nickname: str
    score: int
    rank: int


class SessionResultsUpdate(BaseModel):
    event_id: str
    finish_reason: str
    finished_at: datetime
    participants: list[ParticipantResult]
