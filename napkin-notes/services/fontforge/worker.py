"""Fontforge worker for Napkin Notes font generation."""

import logging
import sys
import time
from typing import Optional

import psycopg2
import redis

from config import Config
from font_generator import FontGenerator
from redis_queue import FontJob, RedisQueue

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)


def ping() -> str:
    """Health check function."""
    return "pong"


def update_font_status(
    font_id: str, status: str, file_path: Optional[str] = None
) -> None:
    """Update font status in PostgreSQL database.

    Args:
        font_id: The font record ID.
        status: New status (processing, ready, failed).
        file_path: Path to the generated font file (set when ready).
    """
    conn = psycopg2.connect(
        host=Config.DB_HOST,
        port=Config.DB_PORT,
        dbname=Config.DB_NAME,
        user=Config.DB_USER,
        password=Config.DB_PASSWORD,
    )
    try:
        with conn.cursor() as cur:
            if file_path:
                cur.execute(
                    "UPDATE fonts SET status = %s, file_path = %s, updated_at = NOW() "
                    "WHERE id = %s",
                    (status, file_path, font_id),
                )
            else:
                cur.execute(
                    "UPDATE fonts SET status = %s, updated_at = NOW() WHERE id = %s",
                    (status, font_id),
                )
        conn.commit()
    finally:
        conn.close()


def process_job(job: FontJob) -> None:
    """Process a single font generation job.

    Args:
        job: The font generation job to process.
    """
    logger.info(f"Processing job for font_id={job.font_id}")
    update_font_status(job.font_id, "processing", None)

    generator = FontGenerator()
    success = generator.generate(job.template_scan_path, job.output_path)

    if success:
        logger.info(f"Font generated successfully: {job.output_path}")
        update_font_status(job.font_id, "ready", job.output_path)
    else:
        logger.error(f"Font generation failed for font_id={job.font_id}")
        update_font_status(job.font_id, "failed", None)


def main() -> None:
    """Main worker loop: connect to Redis and process font generation jobs."""
    logger.info("Starting font generation worker...")
    logger.info(f"Redis URL: {Config.REDIS_URL}")
    logger.info(f"Queue: {Config.QUEUE_NAME}")

    client = redis.from_url(Config.REDIS_URL)
    queue = RedisQueue(client=client, queue_name=Config.QUEUE_NAME)

    logger.info("Worker ready, waiting for jobs...")
    while True:
        try:
            job = queue.dequeue(timeout=5)
            if job is not None:
                process_job(job)
        except KeyboardInterrupt:
            logger.info("Worker shutting down...")
            break
        except Exception as e:
            logger.error(f"Error processing job: {e}")
            time.sleep(1)


if __name__ == "__main__":
    main()
