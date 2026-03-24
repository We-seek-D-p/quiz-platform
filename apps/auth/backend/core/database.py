from typing import AsyncGenerator
from sqlmodel.ext.asyncio.session import AsyncSession
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker
from backend.core.config import settings

async_db_url = settings.DATABASE_URL

async_engine = create_async_engine(async_db_url, echo=True, future=True)

AsyncSessionLocal = async_sessionmaker(
    bind=async_engine, class_=AsyncSession, expire_on_commit=False
)


async def get_session() -> AsyncGenerator[AsyncSession | None]:
    async with AsyncSessionLocal() as session:
        yield session
