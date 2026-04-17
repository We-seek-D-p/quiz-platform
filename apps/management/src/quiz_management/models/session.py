from datetime import UTC, datetime
from uuid import UUID, uuid7

from sqlmodel import Field, Relationship, SQLModel

from quiz_management.models.quiz import Quiz


def get_utc_now():
    return datetime.now(UTC).replace(tzinfo=None)


class GameSession(SQLModel, table=True):
    __tablename__ = "game_sessions"
    __table_args__ = {"schema": "management"}
    id: UUID = Field(default_factory=uuid7, primary_key=True)
    quiz_id: UUID = Field(foreign_key="management.quizzes.id")
    room_code: str = Field(max_length=10)
    host_id: UUID
    status: str

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
