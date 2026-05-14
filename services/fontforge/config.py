"""Configuration for the font generation worker."""

import os


class Config:
    """Worker configuration from environment variables."""

    REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379")
    QUEUE_NAME = os.getenv("QUEUE_NAME", "font_generation")
    STORAGE_PATH = os.getenv("STORAGE_PATH", "./storage")
    DB_HOST = os.getenv("DB_HOST", "localhost")
    DB_PORT = os.getenv("DB_PORT", "5432")
    DB_NAME = os.getenv("DB_NAME", "napkin_notes")
    DB_USER = os.getenv("DB_USER", "postgres")
    DB_PASSWORD = os.getenv("DB_PASSWORD", "postgres")
