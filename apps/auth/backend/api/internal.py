from fastapi import APIRouter, Depends, HTTPException, status

from backend.core.dependencies import get_current_user
from backend.models.users import User


router = APIRouter(tags=["internal"])


@router.get("/validate")
async def validate(_: User = Depends(get_current_user)):
    raise HTTPException(
        status_code=status.HTTP_501_NOT_IMPLEMENTED,
        detail="Validate endpoint not implemented yet",
    )
