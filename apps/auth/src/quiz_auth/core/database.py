from collections.abc import AsyncGenerator

from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine
from sqlmodel.ext.asyncio.session import AsyncSession

from quiz_auth.core.config import settings

async_db_url = settings.database_url

async_engine = create_async_engine(async_db_url, echo=True, future=True)

AsyncSessionLocal = async_sessionmaker(
    bind=async_engine, class_=AsyncSession, expire_on_commit=False
)


async def get_session() -> AsyncGenerator[AsyncSession | None]:
    async with AsyncSessionLocal() as session:
        yield session
