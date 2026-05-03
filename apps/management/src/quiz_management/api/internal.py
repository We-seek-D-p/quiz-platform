from fastapi import APIRouter, Depends

from quiz_management.core.dependencies import verify_internal_auth

router = APIRouter(
    prefix="/internal/v1/sessions", tags=["Internal"], dependencies=[Depends(verify_internal_auth)]
)
