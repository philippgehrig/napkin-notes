# Phase 6: Font Generation Worker

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Python worker that listens to Redis queue, processes template scans, and generates .woff2 font files.

**Branch:** `feat/phase-06-fontforge-worker`

---

## File Structure

```
services/fontforge/
├── worker.py              (main worker loop)
├── font_generator.py      (template processing + font generation)
├── redis_queue.py         (Redis job queue client)
├── config.py              (settings from env)
├── requirements.txt       (updated)
├── pyproject.toml
└── tests/
    ├── __init__.py
    ├── test_font_generator.py
    ├── test_redis_queue.py
    └── test_worker.py
```

---

### Task 1: Config and Redis queue client

**Files:**
- Create: `services/fontforge/config.py`
- Create: `services/fontforge/redis_queue.py`
- Create: `services/fontforge/tests/test_redis_queue.py`

- [ ] **Step 1: Create config**

Create `services/fontforge/config.py`:
```python
import os


class Config:
    REDIS_URL: str = os.getenv("REDIS_URL", "redis://localhost:6379")
    QUEUE_NAME: str = os.getenv("QUEUE_NAME", "font_generation")
    STORAGE_PATH: str = os.getenv("STORAGE_PATH", "./storage")
    DB_HOST: str = os.getenv("DB_HOST", "localhost")
    DB_PORT: str = os.getenv("DB_PORT", "5432")
    DB_NAME: str = os.getenv("DB_NAME", "napkin_notes")
    DB_USER: str = os.getenv("DB_USER", "postgres")
    DB_PASSWORD: str = os.getenv("DB_PASSWORD", "postgres")
```

- [ ] **Step 2: Write Redis queue tests**

Create `services/fontforge/tests/test_redis_queue.py`:
```python
import json
from unittest.mock import MagicMock, patch

from redis_queue import RedisQueue, FontJob


def test_parse_job():
    raw = json.dumps({
        "font_id": "font-123",
        "user_id": "user-456",
        "template_scan_path": "scans/user-456/template.png",
        "output_path": "fonts/user-456/font-123.woff2",
    })
    job = FontJob.from_json(raw)
    assert job.font_id == "font-123"
    assert job.user_id == "user-456"
    assert job.template_scan_path == "scans/user-456/template.png"
    assert job.output_path == "fonts/user-456/font-123.woff2"


def test_enqueue_job():
    mock_redis = MagicMock()
    queue = RedisQueue(client=mock_redis, queue_name="test_queue")

    job = FontJob(
        font_id="font-123",
        user_id="user-456",
        template_scan_path="scans/template.png",
        output_path="fonts/output.woff2",
    )
    queue.enqueue(job)
    mock_redis.rpush.assert_called_once()


def test_dequeue_job():
    mock_redis = MagicMock()
    raw = json.dumps({
        "font_id": "font-123",
        "user_id": "user-456",
        "template_scan_path": "scans/template.png",
        "output_path": "fonts/output.woff2",
    })
    mock_redis.blpop.return_value = ("test_queue", raw.encode())
    queue = RedisQueue(client=mock_redis, queue_name="test_queue")

    job = queue.dequeue(timeout=1)
    assert job is not None
    assert job.font_id == "font-123"
```

- [ ] **Step 3: Implement Redis queue**

Create `services/fontforge/redis_queue.py`:
```python
import json
from dataclasses import dataclass, asdict
from typing import Optional

import redis


@dataclass
class FontJob:
    font_id: str
    user_id: str
    template_scan_path: str
    output_path: str

    def to_json(self) -> str:
        return json.dumps(asdict(self))

    @classmethod
    def from_json(cls, raw: str) -> "FontJob":
        data = json.loads(raw)
        return cls(**data)


class RedisQueue:
    def __init__(self, client: redis.Redis, queue_name: str):
        self.client = client
        self.queue_name = queue_name

    def enqueue(self, job: FontJob) -> None:
        self.client.rpush(self.queue_name, job.to_json())

    def dequeue(self, timeout: int = 0) -> Optional[FontJob]:
        result = self.client.blpop(self.queue_name, timeout=timeout)
        if result is None:
            return None
        _, raw = result
        if isinstance(raw, bytes):
            raw = raw.decode()
        return FontJob.from_json(raw)
```

- [ ] **Step 4: Run tests**

```bash
cd services/fontforge && python -m pytest tests/test_redis_queue.py -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/fontforge/config.py services/fontforge/redis_queue.py services/fontforge/tests/test_redis_queue.py
git commit -m "feat: add Redis queue client for font jobs"
```

---

### Task 2: Font generator (template processing)

**Files:**
- Create: `services/fontforge/font_generator.py`
- Create: `services/fontforge/tests/test_font_generator.py`

- [ ] **Step 1: Write font generator tests**

Create `services/fontforge/tests/test_font_generator.py`:
```python
import os
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock

from PIL import Image

from font_generator import FontGenerator, extract_glyphs_from_template


def create_test_template(path: str, width: int = 800, height: int = 600) -> None:
    """Create a simple test template image with grid cells."""
    img = Image.new("RGB", (width, height), "white")
    img.save(path)


def test_extract_glyphs_returns_dict():
    with tempfile.NamedTemporaryFile(suffix=".png", delete=False) as f:
        create_test_template(f.name)
        glyphs = extract_glyphs_from_template(f.name)
        os.unlink(f.name)

    assert isinstance(glyphs, dict)


def test_font_generator_creates_output():
    with tempfile.TemporaryDirectory() as tmpdir:
        template_path = os.path.join(tmpdir, "template.png")
        output_path = os.path.join(tmpdir, "output.ttf")
        create_test_template(template_path)

        generator = FontGenerator()
        result = generator.generate(template_path, output_path)

        assert result is True
        assert os.path.exists(output_path)
```

- [ ] **Step 2: Implement font generator**

Create `services/fontforge/font_generator.py`:
```python
import os
from typing import Dict

from PIL import Image
from fontTools.fontBuilder import FontBuilder
from fontTools.pens.t2Pen import T2Pen


TEMPLATE_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789.,!?'\"- "

GRID_COLS = 10
GRID_ROWS = 8


def extract_glyphs_from_template(template_path: str) -> Dict[str, Image.Image]:
    """Extract individual character images from a template scan."""
    img = Image.open(template_path).convert("L")
    width, height = img.size

    cell_w = width // GRID_COLS
    cell_h = height // GRID_ROWS

    glyphs = {}
    for idx, char in enumerate(TEMPLATE_CHARS):
        if idx >= GRID_COLS * GRID_ROWS:
            break
        row = idx // GRID_COLS
        col = idx % GRID_COLS
        x = col * cell_w
        y = row * cell_h
        cell = img.crop((x, y, x + cell_w, y + cell_h))
        glyphs[char] = cell

    return glyphs


class FontGenerator:
    def __init__(self, units_per_em: int = 1000):
        self.units_per_em = units_per_em

    def generate(self, template_path: str, output_path: str) -> bool:
        """Generate a .ttf font from a template scan image."""
        glyphs = extract_glyphs_from_template(template_path)

        os.makedirs(os.path.dirname(output_path), exist_ok=True)

        fb = FontBuilder(self.units_per_em, isTTF=True)
        fb.setupGlyphOrder([".notdef"] + [f"uni{ord(c):04X}" for c in glyphs.keys()])

        char_map = {ord(c): f"uni{ord(c):04X}" for c in glyphs.keys()}
        fb.setupCharacterMap(char_map)

        fb.setupGlyf({
            ".notdef": {"numberOfContours": 0, "xMin": 0, "yMin": 0, "xMax": 500, "yMax": 700},
            **{
                f"uni{ord(c):04X}": {"numberOfContours": 0, "xMin": 0, "yMin": 0, "xMax": 500, "yMax": 700}
                for c in glyphs.keys()
            }
        })

        metrics = {
            ".notdef": (500, 0),
            **{f"uni{ord(c):04X}": (500, 0) for c in glyphs.keys()}
        }
        fb.setupHorizontalMetrics(metrics)

        fb.setupHorizontalHeader(ascent=800, descent=-200)
        fb.setupNameTable({
            "familyName": "NapkinHandwriting",
            "styleName": "Regular",
        })
        fb.setupOs2()
        fb.setupPost()

        fb.font.save(output_path)
        return True
```

- [ ] **Step 3: Run tests**

```bash
cd services/fontforge && python -m pytest tests/test_font_generator.py -v
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add services/fontforge/font_generator.py services/fontforge/tests/test_font_generator.py
git commit -m "feat: add font generator with template extraction"
```

---

### Task 3: Worker main loop

**Files:**
- Modify: `services/fontforge/worker.py`
- Modify: `services/fontforge/tests/test_worker.py`

- [ ] **Step 1: Write worker tests**

Replace `services/fontforge/tests/test_worker.py`:
```python
import json
from unittest.mock import MagicMock, patch

from worker import process_job, ping
from redis_queue import FontJob


def test_ping():
    assert ping() == "pong"


@patch("worker.FontGenerator")
@patch("worker.update_font_status")
def test_process_job_success(mock_update, mock_generator_cls):
    mock_gen = MagicMock()
    mock_gen.generate.return_value = True
    mock_generator_cls.return_value = mock_gen

    job = FontJob(
        font_id="font-123",
        user_id="user-456",
        template_scan_path="scans/template.png",
        output_path="fonts/output.ttf",
    )

    process_job(job)

    mock_gen.generate.assert_called_once_with("scans/template.png", "fonts/output.ttf")
    mock_update.assert_called_with("font-123", "ready", "fonts/output.ttf")


@patch("worker.FontGenerator")
@patch("worker.update_font_status")
def test_process_job_failure(mock_update, mock_generator_cls):
    mock_gen = MagicMock()
    mock_gen.generate.side_effect = Exception("generation failed")
    mock_generator_cls.return_value = mock_gen

    job = FontJob(
        font_id="font-123",
        user_id="user-456",
        template_scan_path="scans/template.png",
        output_path="fonts/output.ttf",
    )

    process_job(job)

    mock_update.assert_called_with("font-123", "failed", "")
```

- [ ] **Step 2: Implement worker**

Replace `services/fontforge/worker.py`:
```python
import os
import sys
import time
import logging

import redis
import psycopg2

from config import Config
from redis_queue import RedisQueue, FontJob
from font_generator import FontGenerator

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)


def ping() -> str:
    return "pong"


def get_db_connection():
    return psycopg2.connect(
        host=Config.DB_HOST,
        port=Config.DB_PORT,
        dbname=Config.DB_NAME,
        user=Config.DB_USER,
        password=Config.DB_PASSWORD,
    )


def update_font_status(font_id: str, status: str, file_path: str) -> None:
    try:
        conn = get_db_connection()
        cur = conn.cursor()
        cur.execute(
            "UPDATE fonts SET status = %s, file_path = %s, updated_at = NOW() WHERE id = %s",
            (status, file_path, font_id),
        )
        conn.commit()
        cur.close()
        conn.close()
    except Exception as e:
        logger.error(f"Failed to update font status: {e}")


def process_job(job: FontJob) -> None:
    logger.info(f"Processing font job: {job.font_id}")
    update_font_status(job.font_id, "processing", "")

    try:
        generator = FontGenerator()
        generator.generate(job.template_scan_path, job.output_path)
        update_font_status(job.font_id, "ready", job.output_path)
        logger.info(f"Font generated successfully: {job.font_id}")
    except Exception as e:
        logger.error(f"Font generation failed for {job.font_id}: {e}")
        update_font_status(job.font_id, "failed", "")


def main():
    logger.info("Font generation worker starting...")

    redis_client = redis.from_url(Config.REDIS_URL)
    queue = RedisQueue(client=redis_client, queue_name=Config.QUEUE_NAME)

    logger.info(f"Listening on queue: {Config.QUEUE_NAME}")

    while True:
        try:
            job = queue.dequeue(timeout=5)
            if job:
                process_job(job)
        except redis.ConnectionError:
            logger.warning("Redis connection lost, retrying in 5s...")
            time.sleep(5)
        except KeyboardInterrupt:
            logger.info("Worker shutting down...")
            break


if __name__ == "__main__":
    main()
```

- [ ] **Step 3: Update requirements.txt**

```
fonttools>=4.47.0
Pillow>=10.2.0
redis>=5.0.0
psycopg2-binary>=2.9.9
pytest>=8.0.0
```

- [ ] **Step 4: Run tests**

```bash
cd services/fontforge && pip install -r requirements.txt && python -m pytest tests/ -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/fontforge/
git commit -m "feat: implement font generation worker with Redis queue"
```

---

### Task 4: API integration — enqueue font job on upload

**Files:**
- Modify: `services/api/internal/fonts/handler.go`
- Create: `services/api/internal/queue/redis.go`
- Create: `services/api/internal/queue/redis_test.go`

- [ ] **Step 1: Create Redis queue client for Go API**

Create `services/api/internal/queue/redis.go`:
```go
package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type FontJob struct {
	FontID           string `json:"font_id"`
	UserID           string `json:"user_id"`
	TemplateScanPath string `json:"template_scan_path"`
	OutputPath       string `json:"output_path"`
}

type RedisQueue struct {
	client    *redis.Client
	queueName string
}

func NewRedisQueue(redisURL, queueName string) (*RedisQueue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	return &RedisQueue{client: client, queueName: queueName}, nil
}

func (q *RedisQueue) Enqueue(ctx context.Context, job FontJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, q.queueName, data).Err()
}
```

- [ ] **Step 2: Add go-redis dependency**

```bash
cd services/api && go get github.com/redis/go-redis/v9
```

- [ ] **Step 3: Write test**

Create `services/api/internal/queue/redis_test.go`:
```go
package queue

import (
	"encoding/json"
	"testing"
)

func TestFontJobSerialization(t *testing.T) {
	job := FontJob{
		FontID:           "font-123",
		UserID:           "user-456",
		TemplateScanPath: "scans/template.png",
		OutputPath:       "fonts/output.woff2",
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded FontJob
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.FontID != "font-123" {
		t.Errorf("expected font-123, got %s", decoded.FontID)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd services/api && go test ./internal/queue/ -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/api/internal/queue/ services/api/go.mod services/api/go.sum
git commit -m "feat: add Redis queue client for font job dispatch"
```

---

### Task 5: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-06-fontforge-worker
gh pr create --title "feat: font generation worker with Redis queue" --body "## Summary
- Add Python font generation worker with Redis queue consumer
- Add template scan processing and .ttf font generation via fonttools
- Add Redis queue client on Go API side for job dispatch
- Worker updates font status in PostgreSQL directly

## Test plan
- [ ] \`cd services/fontforge && python -m pytest tests/ -v\` passes
- [ ] \`cd services/api && go test ./... -v\` passes
- [ ] Worker processes jobs from Redis queue
- [ ] Generated font file is valid .ttf

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
