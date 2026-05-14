"""Tests for the font generator module."""

import os
import tempfile

from PIL import Image

from font_generator import FontGenerator, extract_glyphs_from_template


class TestExtractGlyphs:
    """Tests for glyph extraction from template images."""

    def test_extract_returns_dict(self):
        """extract_glyphs_from_template returns a dict of char to Image."""
        # Create a minimal template image (10 cols x 8 rows grid)
        width = 1000
        height = 800
        img = Image.new("L", (width, height), color=255)

        with tempfile.NamedTemporaryFile(suffix=".png", delete=False) as f:
            img.save(f.name)
            template_path = f.name

        try:
            glyphs = extract_glyphs_from_template(template_path)
            assert isinstance(glyphs, dict)
            # Should have entries (up to 80 cells in 10x8 grid)
            assert len(glyphs) > 0
            # Each value should be a PIL Image
            for char, glyph_img in glyphs.items():
                assert isinstance(char, str)
                assert isinstance(glyph_img, Image.Image)
        finally:
            os.unlink(template_path)

    def test_extract_glyph_dimensions(self):
        """Each extracted glyph has expected cell dimensions."""
        width = 1000
        height = 800
        img = Image.new("L", (width, height), color=255)

        with tempfile.NamedTemporaryFile(suffix=".png", delete=False) as f:
            img.save(f.name)
            template_path = f.name

        try:
            glyphs = extract_glyphs_from_template(template_path)
            cell_width = width // 10
            cell_height = height // 8
            for _, glyph_img in glyphs.items():
                assert glyph_img.size == (cell_width, cell_height)
        finally:
            os.unlink(template_path)


class TestFontGenerator:
    """Tests for FontGenerator class."""

    def test_generate_creates_output_file(self):
        """FontGenerator.generate creates a .ttf file at output_path."""
        # Create a template image
        width = 1000
        height = 800
        img = Image.new("L", (width, height), color=255)

        with tempfile.NamedTemporaryFile(suffix=".png", delete=False) as f:
            img.save(f.name)
            template_path = f.name

        with tempfile.NamedTemporaryFile(suffix=".ttf", delete=False) as f:
            output_path = f.name

        try:
            generator = FontGenerator()
            result = generator.generate(template_path, output_path)
            assert result is True
            assert os.path.exists(output_path)
            assert os.path.getsize(output_path) > 0
        finally:
            os.unlink(template_path)
            if os.path.exists(output_path):
                os.unlink(output_path)

    def test_generate_returns_false_on_invalid_template(self):
        """FontGenerator.generate returns False for non-existent template."""
        with tempfile.NamedTemporaryFile(suffix=".ttf", delete=False) as f:
            output_path = f.name

        try:
            generator = FontGenerator()
            result = generator.generate("/nonexistent/template.png", output_path)
            assert result is False
        finally:
            if os.path.exists(output_path):
                os.unlink(output_path)
