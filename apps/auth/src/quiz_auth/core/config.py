from pathlib import Path

from pydantic import field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict

BASE_DIR = Path(__file__).resolve().parents[3]


class Settings(BaseSettings):
    app_name: str = "Quiz Auth"
    debug: bool = False

    database_url: str

    jwt_secret_key: str
    jwt_algorithm: str = "HS256"
    access_token_ttl_minutes: int = 5
    refresh_token_ttl_days: int = 30
    refresh_cookie_secure: bool = True

    model_config = SettingsConfigDict(
        env_prefix="AUTH_",
        env_file=BASE_DIR / ".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    @field_validator("jwt_secret_key")
    @classmethod
    def validate_jwt_secret_key(cls, value: str) -> str:
        if len(value) < 32:
            raise ValueError("AUTH_JWT_SECRET_KEY must be at least 32 characters long")
        return value


settings = Settings()
