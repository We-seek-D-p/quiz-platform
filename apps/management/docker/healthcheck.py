import sys
from http.client import HTTPConnection, HTTPSConnection
from urllib.parse import urlparse


def main() -> int:
    if len(sys.argv) != 2:
        return 1

    parsed = urlparse(sys.argv[1])
    if parsed.scheme not in {"http", "https"}:
        return 1
    if not parsed.hostname:
        return 1

    port = parsed.port or (443 if parsed.scheme == "https" else 80)
    target = parsed.path or "/"
    if parsed.query:
        target = f"{target}?{parsed.query}"

    try:
        connection_class = HTTPSConnection if parsed.scheme == "https" else HTTPConnection
        connection = connection_class(parsed.hostname, port, timeout=2)
        connection.request("GET", target)
        response = connection.getresponse()
        status = response.status
        connection.close()
        return 0 if 200 <= status < 400 else 1
    except OSError:
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
