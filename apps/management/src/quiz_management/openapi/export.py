import json
from pathlib import Path

from quiz_management.main import app


def export_openapi() -> Path:
    output_path = Path(__file__).resolve().parents[3] / "openapi" / "openapi.json"
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(app.openapi(), indent=2) + "\n")
    return output_path


if __name__ == "__main__":
    export_openapi()
