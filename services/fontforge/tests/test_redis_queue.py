"""Tests for the Redis queue client."""

import json
from unittest.mock import MagicMock, patch

from redis_queue import FontJob, RedisQueue


class TestFontJob:
    """Tests for FontJob dataclass serialization."""

    def test_parse_job(self):
        """FontJob can be deserialized from JSON."""
        data = json.dumps({
            "font_id": "abc-123",
            "user_id": "user-456",
            "template_scan_path": "/storage/scans/template.png",
            "output_path": "/storage/fonts/output.ttf",
        })
        job = FontJob.from_json(data)
        assert job.font_id == "abc-123"
        assert job.user_id == "user-456"
        assert job.template_scan_path == "/storage/scans/template.png"
        assert job.output_path == "/storage/fonts/output.ttf"

    def test_to_json(self):
        """FontJob can be serialized to JSON."""
        job = FontJob(
            font_id="abc-123",
            user_id="user-456",
            template_scan_path="/storage/scans/template.png",
            output_path="/storage/fonts/output.ttf",
        )
        data = json.loads(job.to_json())
        assert data["font_id"] == "abc-123"
        assert data["user_id"] == "user-456"
        assert data["template_scan_path"] == "/storage/scans/template.png"
        assert data["output_path"] == "/storage/fonts/output.ttf"


class TestRedisQueue:
    """Tests for RedisQueue enqueue/dequeue operations."""

    def test_enqueue_job(self):
        """Enqueue pushes job JSON to the Redis list."""
        mock_redis = MagicMock()
        queue = RedisQueue(client=mock_redis, queue_name="test_queue")

        job = FontJob(
            font_id="abc-123",
            user_id="user-456",
            template_scan_path="/scans/template.png",
            output_path="/fonts/output.ttf",
        )
        queue.enqueue(job)

        mock_redis.lpush.assert_called_once_with("test_queue", job.to_json())

    def test_dequeue_job(self):
        """Dequeue pops and parses a job from the Redis list."""
        mock_redis = MagicMock()
        queue = RedisQueue(client=mock_redis, queue_name="test_queue")

        job_data = json.dumps({
            "font_id": "abc-123",
            "user_id": "user-456",
            "template_scan_path": "/scans/template.png",
            "output_path": "/fonts/output.ttf",
        })
        mock_redis.brpop.return_value = ("test_queue", job_data)

        job = queue.dequeue(timeout=5)

        assert job is not None
        assert job.font_id == "abc-123"
        assert job.user_id == "user-456"
        mock_redis.brpop.assert_called_once_with("test_queue", timeout=5)

    def test_dequeue_returns_none_on_timeout(self):
        """Dequeue returns None when Redis times out."""
        mock_redis = MagicMock()
        queue = RedisQueue(client=mock_redis, queue_name="test_queue")

        mock_redis.brpop.return_value = None

        job = queue.dequeue(timeout=1)

        assert job is None
