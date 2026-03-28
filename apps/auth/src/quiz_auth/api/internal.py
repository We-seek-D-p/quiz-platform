from fastapi import APIRouter, Depends, Response, status

from quiz_auth.core.dependencies import get_current_user
from quiz_auth.models.users import User


router = APIRouter(tags=["internal"])


@router.get("/validate", status_code=status.HTTP_200_OK)
async def validate(current_user: User = Depends(get_current_user)) -> Response:
    response = Response(status_code=status.HTTP_200_OK)
    response.headers["X-Auth-User-Id"] = str(current_user.id)
    response.headers["X-Auth-Role"] = current_user.role
    return response
