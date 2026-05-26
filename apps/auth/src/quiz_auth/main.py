from fastapi import FastAPI

from quiz_auth.api.auth import router as auth_router
from quiz_auth.api.internal import router as internal_router
from quiz_auth.api.users import router as user_router

app = FastAPI(title="Quiz App")


@app.get("/livez", include_in_schema=False)
async def livez():
    return {"status": "ok"}


@app.get("/readyz", include_in_schema=False)
async def readyz():
    return {"status": "ready"}


app.include_router(auth_router, prefix="/api/v1")
app.include_router(user_router, prefix="/api/v1")
app.include_router(internal_router, prefix="/api/v1")
