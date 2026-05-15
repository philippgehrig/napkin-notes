"""Tests for the font generation worker."""

from unittest.mock import MagicMock, patch

from redis_queue import FontJob
from worker import ping, process_job


class TestPing:
    """Tests for health check."""

    def test_ping(self):
        """ping returns pong."""
        assert ping() == "pong"


class TestProcessJob:
    """Tests for job processing logic."""

    @patch("worker.update_font_status")
    @patch("worker.FontGenerator")
    def test_process_job_success(self, mock_generator_cls, mock_update_status):
        """process_job calls generator and updates status to ready on success."""
        mock_generator = MagicMock()
        mock_generator.generate.return_value = True
        mock_generator_cls.return_value = mock_generator

        job = FontJob(
            font_id="font-123",
            user_id="user-456",
            template_scan_path="/storage/scans/template.png",
            output_path="/storage/fonts/output.ttf",
        )

        process_job(job)

        # Should update to processing first, then to ready
        calls = mock_update_status.call_args_list
        assert len(calls) == 2
        assert calls[0].args == ("font-123", "processing", None)
        assert calls[1].args == ("font-123", "ready", "/storage/fonts/output.ttf")

        mock_generator.generate.assert_called_once_with(
            "/storage/scans/template.png", "/storage/fonts/output.ttf"
        )

    @patch("worker.update_font_status")
    @patch("worker.FontGenerator")
    def test_process_job_failure(self, mock_generator_cls, mock_update_status):
        """process_job updates status to failed when generator fails."""
        mock_generator = MagicMock()
        mock_generator.generate.return_value = False
        mock_generator_cls.return_value = mock_generator

        job = FontJob(
            font_id="font-123",
            user_id="user-456",
            template_scan_path="/storage/scans/template.png",
            output_path="/storage/fonts/output.ttf",
        )

        process_job(job)

        # Should update to processing first, then to failed
        calls = mock_update_status.call_args_list
        assert len(calls) == 2
        assert calls[0].args == ("font-123", "processing", None)
        assert calls[1].args == ("font-123", "failed", None)
