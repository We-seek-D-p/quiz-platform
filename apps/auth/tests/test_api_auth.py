from unittest.mock import AsyncMock, patch
from uuid import uuid7
from datetime import datetime, timedelta, timezone


class MockTockenPair:
    def __init__(self):
        now_utc = datetime.now(timezone.utc)
        self.access_token = 'access_token-123'
        self.refresh_token = 'refresh_token-456'
        self.token_type = 'Bearer'
        self.access_expires_in = 900
        self.refresh_expires_in = 604800
        self.access_token_expires_at = now_utc + timedelta(minutes=15)
        self.refresh_token_expires_at = now_utc + timedelta(weeks=1)
        self.session_id = uuid7()
