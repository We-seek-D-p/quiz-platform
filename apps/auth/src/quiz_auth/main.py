from fastapi import FastAPI

from quiz_auth.api.auth import router as auth_router
from quiz_auth.api.internal import router as internal_router
from quiz_auth.api.users import router as user_router


app = FastAPI(title="Quiz App")

app.include_router(auth_router)

app.include_router(user_router)
app.include_router(internal_router)
