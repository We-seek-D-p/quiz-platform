from sqlmodel import select
from sqlmodel.ext.asyncio.session import AsyncSession

from backend.models.users import Role


class RoleRepository:
    def __init__(self, db: AsyncSession):
        self.db = db

    async def get_by_slug(self, slug: str) -> Role | None:
        result = await self.db.exec(select(Role).where(Role.slug == slug))
        return result.first()

    async def ensure_host_role(self) -> Role:
        role = await self.get_by_slug("host")
        if role:
            return role
        role = Role(slug="host", name="Host", priority=0)
        self.db.add(role)
        await self.db.commit()
        await self.db.refresh(role)
        return role
