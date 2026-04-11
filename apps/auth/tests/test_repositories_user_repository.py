from unittest.mock import AsyncMock
from uuid import uuid7

import pytest
from fastapi import HTTPException


def test_update_user_raises_404_when_user_missing(user_repository, run_async) -> None:
    user_repository.get_by_id = AsyncMock(return_value=None)

    with pytest.raises(HTTPException) as exc:
        run_async(user_repository.update_user(uuid7(), {"nickname": "bob"}))

    assert exc.value.status_code == 404
    assert exc.value.detail == "User not found"
