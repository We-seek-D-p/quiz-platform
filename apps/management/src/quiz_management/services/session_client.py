from datetime import UTC, datetime
from uuid import UUID

import httpx

from quiz_management.core.config import settings


class SessionServiceClient:
    def __init__(self):
        self.base_url = settings.session_service_url
        self.headers = {
            "X-Internal-Service": settings.management_service_name,
            "X-Internal-Token": settings.management_internal_token,
        }

    async def init_session(
        self, session_id: UUID, quiz_id: UUID, host_id: UUID, idempotency_key: str
    ) -> httpx.Response:
        url = f"{self.base_url}/internal/v1/sessions/{session_id}"
        payload = {
            "quiz_id": str(quiz_id),
            "host_id": str(host_id),
            "created_at": datetime.now(UTC).isoformat(),
        }
        async with httpx.AsyncClient(timeout=settings.session_service_timeout) as client:
            return await client.put(
                url, json=payload, headers={**self.headers, "Idempotency-Key": idempotency_key}
            )

    async def get_session(self, session_id: UUID) -> httpx.Response:
        url = f"{self.base_url}/internal/v1/sessions/{session_id}"
        async with httpx.AsyncClient(timeout=settings.session_service_timeout) as client:
            return await client.get(url, headers=self.headers)

    async def delete_session(self, session_id: UUID) -> httpx.Response:
        url = f"{self.base_url}/internal/v1/sessions/{session_id}"
        async with httpx.AsyncClient(timeout=settings.session_service_timeout) as client:
            return await client.delete(url, headers=self.headers)
