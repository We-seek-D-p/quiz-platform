from pathlib import Path

from pydantic import AliasChoices, Field
from pydantic_settings import BaseSettings, SettingsConfigDict

BASE_DIR = Path(__file__).resolve().parents[3]


class Settings(BaseSettings):
    app_name: str = "Quiz Management"
    debug: bool = False

    database_url: str

    internal_service_name: str = "management"
    internal_allowed_services: str = "session"
    internal_token: str = "placeholder_token"  # noqa: S105
    session_internal_token: str = Field(
        default="placeholder_token", validation_alias=AliasChoices("MANAGEMENT_SESSION_INTERNAL_TOKEN", "SESSION_INTERNAL_TOKEN")
    )
    session_service_url: str = Field(
        default="http://session:8000",
        validation_alias=AliasChoices("MANAGEMENT_SESSION_SERVICE_URL", "SESSION_MANAGEMENT_BASE_URL"),
    )

    model_config = SettingsConfigDict(
        env_prefix="MANAGEMENT_",
        env_file=BASE_DIR / ".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )


settings = Settings()
