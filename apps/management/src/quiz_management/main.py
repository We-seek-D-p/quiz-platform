from fastapi import FastAPI

from quiz_management.api.question import router as question_router
from quiz_management.api.quiz import router as quiz_router
from quiz_management.api.session import router as session_router

app = FastAPI(title="Quiz Management")

app.include_router(quiz_router)
app.include_router(question_router)
app.include_router(session_router)
