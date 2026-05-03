from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse

from quiz_management.api.question import router as question_router
from quiz_management.api.quiz import router as quiz_router
from quiz_management.api.session import router as session_router

app = FastAPI(title="Quiz Management")

app.include_router(quiz_router, prefix="/api/v1")
app.include_router(question_router, prefix="/api/v1")
app.include_router(session_router, prefix="/api/v1")

STATUS_TO_CODE = {
    400: "invalid_payload",
    401: "unauthorized",
    403: "forbidden",
    404: "not_found",
    405: "method_not_allowed",
    409: "conflict",
    422: "validation_error",
    429: "too_many_requests",
    500: "internal_error",
    503: "service_unavailable",
}


@app.exception_handler(HTTPException)
async def unified_exception_handler(_: Request, exc: HTTPException):
    if isinstance(exc.detail, dict) and "code" in exc.detail:
        return JSONResponse(status_code=exc.status_code, content=exc.detail)

    code = STATUS_TO_CODE.get(exc.status_code, "error")
    message = str(exc.detail) if exc.detail else "An unexpected error occurred"

    return JSONResponse(status_code=exc.status_code, content={"code": code, "message": message})
