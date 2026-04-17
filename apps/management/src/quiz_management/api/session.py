from typing import Annotated

from fastapi import APIRouter, status
from fastapi.params import Depends

from quiz_management.core.dependencies import get_session_service
from quiz_management.repositories.session_repositories import TSession
from quiz_management.services.session import SessionService

router = APIRouter(prefix="/sessions", tags=["Sessions"])


@router.post("", status_code=status.HTTP_201_CREATED)
async def create_session(
    data: TSession, service: Annotated[SessionService, Depends(get_session_service)]
):
    await service.create_session(data)
