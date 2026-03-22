from fastapi import FastAPI

from backend.api.auth import router as auth_router
from backend.api.internal import router as internal_router


app = FastAPI(title="Quiz App")

app.include_router(auth_router)
app.include_router(internal_router)
