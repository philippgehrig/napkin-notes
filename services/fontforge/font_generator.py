"""Font generator using fontTools to create TTF fonts from template scans."""

import os
import string
from typing import Dict

from fontTools.fontBuilder import FontBuilder
from PIL import Image


# Character map for the 10x8 grid template
TEMPLATE_CHARS = (
    string.ascii_uppercase
    + string.ascii_lowercase
    + string.digits
    + "!@#$%^&*()-_=+[]{}|;:',.<>?/~`"
)


def extract_glyphs_from_template(template_path: str) -> Dict[str, Image.Image]:
    """Open a template image and divide into a 10 col x 8 row grid.

    Returns a dict mapping character to its glyph image.
    """
    img = Image.open(template_path)
    width, height = img.size

    cols = 10
    rows = 8
    cell_width = width // cols
    cell_height = height // rows

    glyphs = {}
    for row in range(rows):
        for col in range(cols):
            idx = row * cols + col
            if idx >= len(TEMPLATE_CHARS):
                break
            char = TEMPLATE_CHARS[idx]
            left = col * cell_width
            top = row * cell_height
            right = left + cell_width
            bottom = top + cell_height
            glyph_img = img.crop((left, top, right, bottom))
            glyphs[char] = glyph_img

    return glyphs


class FontGenerator:
    """Generates TTF fonts from template scan images."""

    def generate(self, template_path: str, output_path: str) -> bool:
        """Generate a TTF font from a template scan.

        Args:
            template_path: Path to the scanned template image.
            output_path: Path where the .ttf file will be written.

        Returns:
            True on success, False on failure.
        """
        try:
            glyphs = extract_glyphs_from_template(template_path)
        except (FileNotFoundError, OSError):
            return False

        try:
            self._build_font(glyphs, output_path)
            return True
        except Exception:
            return False

    def _build_font(
        self, glyphs: Dict[str, Image.Image], output_path: str
    ) -> None:
        """Build a minimal TTF font using fontTools FontBuilder."""
        units_per_em = 1000
        ascent = 800
        descent = -200

        # Glyph names: .notdef, space, plus one per character
        glyph_names = [".notdef", "space"]
        char_to_glyph = {}
        for char in glyphs:
            glyph_name = f"uni{ord(char):04X}"
            glyph_names.append(glyph_name)
            char_to_glyph[char] = glyph_name

        # Character map (cmap)
        cmap = {ord(char): glyph_name for char, glyph_name in char_to_glyph.items()}
        cmap[0x20] = "space"  # space character

        fb = FontBuilder(units_per_em, isTTF=True)
        fb.setupGlyphOrder(glyph_names)
        fb.setupCharacterMap(cmap)

        # Build glyph outlines - simple rectangles as placeholders
        # representing where handwriting strokes would go
        fb.setupGlyf({
            name: self._make_glyph_outline(name, units_per_em, ascent)
            for name in glyph_names
        })

        # Metrics: all glyphs same advance width
        advance_width = 600
        metrics = {name: (advance_width, 0) for name in glyph_names}
        metrics[".notdef"] = (500, 0)
        metrics["space"] = (300, 0)
        fb.setupHorizontalMetrics(metrics)

        fb.setupHorizontalHeader(ascent=ascent, descent=descent)
        fb.setupNameTable({
            "familyName": "NapkinNotes Handwriting",
            "styleName": "Regular",
        })
        fb.setupOS2()
        fb.setupPost()
        fb.setupHead(unitsPerEm=units_per_em)

        # Ensure output directory exists
        os.makedirs(os.path.dirname(output_path) or ".", exist_ok=True)
        fb.font.save(output_path)

    def _make_glyph_outline(self, name: str, units_per_em: int, ascent: int):
        """Create a simple glyph outline (empty for .notdef/space, box otherwise)."""
        from fontTools.pens.ttGlyphPen import TTGlyphPen

        pen = TTGlyphPen(glyphSet=None)

        if name in (".notdef", "space"):
            # Empty glyph - draw nothing
            return pen.glyph()

        # Simple rectangular outline as placeholder
        pen.moveTo((100, 0))
        pen.lineTo((500, 0))
        pen.lineTo((500, 700))
        pen.lineTo((100, 700))
        pen.closePath()
        return pen.glyph()
