from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends

from quiz_management.core.dependencies import get_session_service, verify_internal_auth
from quiz_management.models.session import SessionBootstrap
from quiz_management.services.session import SessionService

router = APIRouter(
    prefix="/internal/v1/sessions", tags=["Internal"], dependencies=[Depends(verify_internal_auth)]
)


@router.get("/{session_id}/bootstrap", response_model=SessionBootstrap)
async def get_bootstrap(
    session_id: UUID,
    service: Annotated[SessionService, Depends(get_session_service)],
):
    session = await service.get_bootstrap_data(session_id)

    return {"session": session, "quiz_snapshot": session.quiz}
