from pathlib import Path

from pydantic_settings import BaseSettings, SettingsConfigDict

BASE_DIR = Path(__file__).resolve().parents[3]


class Settings(BaseSettings):
    app_name: str = "Quiz Management"
    debug: bool = False

    database_url: str

    session_service_url: str = "http://localhost:8000"
    session_internal_token: str = "secret"  # noqa: S105
    session_service_name: str = "management-service"

    model_config = SettingsConfigDict(
        env_prefix="MANAGEMENT_",
        env_file=BASE_DIR / ".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )


settings = Settings()
