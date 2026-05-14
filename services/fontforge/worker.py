"""Fontforge worker for Napkin Notes font generation."""


def ping() -> str:
    """Health check function."""
    return "pong"


if __name__ == "__main__":
    print(ping())
