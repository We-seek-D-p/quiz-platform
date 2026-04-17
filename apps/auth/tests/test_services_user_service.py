from uuid import uuid7

import pytest
from fastapi import HTTPException


def test_get_user_by_id_forbidden_for_other_user(
    user_service, run_async, fake_user_factory
) -> None:
    requested_user_id = uuid7()
    current_user = fake_user_factory()

    with pytest.raises(HTTPException) as exc:
        run_async(user_service.get_user_by_id(requested_user_id, current_user))

    assert exc.value.status_code == 403
    assert exc.value.detail == "Not enough permissions"
