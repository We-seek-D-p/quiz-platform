from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, status

from quiz_management.core.dependencies import get_session_service, verify_internal_auth
from quiz_management.models.session import SessionBootstrap, SessionStatusUpdate
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


@router.patch("/{session_id}/status", status_code=status.HTTP_204_NO_CONTENT)
async def update_session_status(
    session_id: UUID,
    data: SessionStatusUpdate,
    service: Annotated[SessionService, Depends(get_session_service)],
):
    await service.update_session_status(session_id, data)
