"""Redis queue client for font generation jobs."""

import json
from dataclasses import dataclass
from typing import Optional


@dataclass
class FontJob:
    """Represents a font generation job."""

    font_id: str
    user_id: str
    template_scan_path: str
    output_path: str

    def to_json(self) -> str:
        """Serialize job to JSON string."""
        return json.dumps({
            "font_id": self.font_id,
            "user_id": self.user_id,
            "template_scan_path": self.template_scan_path,
            "output_path": self.output_path,
        })

    @classmethod
    def from_json(cls, data: str) -> "FontJob":
        """Deserialize job from JSON string."""
        parsed = json.loads(data)
        return cls(
            font_id=parsed["font_id"],
            user_id=parsed["user_id"],
            template_scan_path=parsed["template_scan_path"],
            output_path=parsed["output_path"],
        )


class RedisQueue:
    """Queue backed by a Redis list."""

    def __init__(self, client, queue_name: str):
        """Initialize with a Redis client and queue name."""
        self.client = client
        self.queue_name = queue_name

    def enqueue(self, job: FontJob) -> None:
        """Push a job onto the queue."""
        self.client.lpush(self.queue_name, job.to_json())

    def dequeue(self, timeout: int = 0) -> Optional[FontJob]:
        """Pop a job from the queue. Returns None on timeout."""
        result = self.client.brpop(self.queue_name, timeout=timeout)
        if result is None:
            return None
        _, data = result
        return FontJob.from_json(data)
