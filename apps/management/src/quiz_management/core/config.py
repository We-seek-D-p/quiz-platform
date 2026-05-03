from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict

BASE_DIR = Path(__file__).resolve().parents[3]


class Settings(BaseSettings):
    app_name: str = "Quiz Management"
    debug: bool = False

    database_url: str

    internal_service_name: str
    internal_allowed_services: str
    internal_token: str
    session_service_url: str = Field(validation_alias="SESSION_MANAGEMENT_BASE_URL")

    model_config = SettingsConfigDict(
        env_prefix="MANAGEMENT_",
        env_file=BASE_DIR / ".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )


settings = Settings()
