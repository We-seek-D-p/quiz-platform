"""make refresh token expiration timestamptz

Revision ID: 4fb1149f4b8a
Revises: da132227e5a8
Create Date: 2026-03-25 00:30:00.000000

"""

from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = "4fb1149f4b8a"
down_revision: Union[str, Sequence[str], None] = "da132227e5a8"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.alter_column(
        "refresh_tokens",
        "expires_at",
        schema="auth",
        existing_type=sa.DateTime(),
        type_=sa.DateTime(timezone=True),
        postgresql_using="expires_at AT TIME ZONE 'UTC'",
    )


def downgrade() -> None:
    op.alter_column(
        "refresh_tokens",
        "expires_at",
        schema="auth",
        existing_type=sa.DateTime(timezone=True),
        type_=sa.DateTime(),
        postgresql_using="timezone('UTC', expires_at)",
    )
