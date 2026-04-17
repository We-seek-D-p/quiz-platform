from hashlib import sha256

from quiz_auth.utils.security import hash_refresh_token


def test_hash_refresh_token_matches_sha256(token_factory) -> None:
    token = token_factory()

    hashed = hash_refresh_token(token)

    assert hashed == sha256(token.encode("utf-8")).hexdigest()
