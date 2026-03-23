from datetime import datetime, timezone
from uuid import UUID, uuid7
from sqlmodel import Field, SQLModel, Relationship

from models.quizz import Quiz


def get_utc_now():
    return datetime.now(timezone.utc)


class GameSession(SQLModel, table=True):
    __tablename__ = "game_sessions"
    id: UUID = Field(default_factory=uuid7, primary_key=True)
    quiz_id: UUID = Field(foreign_key="quizzes.id")
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
    id: UUID = Field(default_factory=uuid7, primary_key=True)
    session_id: UUID = Field(foreign_key="game_sessions.id")
    player_nickname: str = Field(max_length=255)
    score: int = Field(default=0)
    rank: int | None = None

    session: GameSession = Relationship(back_populates="participants")
