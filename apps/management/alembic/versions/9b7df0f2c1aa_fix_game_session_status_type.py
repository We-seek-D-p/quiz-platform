"""fix game session status type

Revision ID: 9b7df0f2c1aa
Revises: 230227e69a40
Create Date: 2026-05-06 06:10:00.000000

"""

from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = "9b7df0f2c1aa"
down_revision: Union[str, Sequence[str], None] = "230227e69a40"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Upgrade schema."""
    op.alter_column(
        "game_sessions",
        "status",
        existing_type=sa.String(),
        type_=sa.String(length=32),
        postgresql_using="status::text",
        schema="management",
    )

    op.execute("DROP TYPE IF EXISTS sessionstatus")


def downgrade() -> None:
    """Downgrade schema."""
    op.execute(
        """
        DO $$ BEGIN
            CREATE TYPE sessionstatus AS ENUM (
                'initializing',
                'lobby',
                'in_progress',
                'finished',
                'init_failed'
            );
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;
        """
    )

    op.alter_column(
        "game_sessions",
        "status",
        existing_type=sa.String(length=32),
        type_=sa.Enum(
            "initializing",
            "lobby",
            "in_progress",
            "finished",
            "init_failed",
            name="sessionstatus",
        ),
        postgresql_using="status::sessionstatus",
        schema="management",
    )
