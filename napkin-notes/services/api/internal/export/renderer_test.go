package export

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRenderNapkin_ValidDimensions(t *testing.T) {
	opts := RenderOptions{
		Content:  "Hello napkin",
		Width:    800,
		Height:   600,
		FontSize: 28,
	}

	img, err := RenderNapkin(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 {
		t.Errorf("expected width=800, got %d", bounds.Dx())
	}
	if bounds.Dy() != 600 {
		t.Errorf("expected height=600, got %d", bounds.Dy())
	}
}

func TestRenderNapkin_EncodesAsPNG(t *testing.T) {
	opts := RenderOptions{
		Content:  "Test PNG encoding",
		Width:    400,
		Height:   300,
		FontSize: 20,
	}

	img, err := RenderNapkin(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode as PNG: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PNG data")
	}
}

func TestRenderNapkin_EmptyContent(t *testing.T) {
	opts := RenderOptions{
		Content:  "",
		Width:    800,
		Height:   600,
		FontSize: 28,
	}

	img, err := RenderNapkin(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 600 {
		t.Errorf("expected 800x600, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestRenderNapkin_BackgroundColor(t *testing.T) {
	opts := RenderOptions{
		Content:  "",
		Width:    100,
		Height:   100,
		FontSize: 14,
	}

	img, err := RenderNapkin(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the top-left pixel is the napkin background color #FFF8E7
	r, g, b, a := img.At(0, 0).RGBA()
	// Convert from 16-bit to 8-bit
	r8, g8, b8, a8 := r>>8, g>>8, b>>8, a>>8

	if r8 != 0xFF || g8 != 0xF8 || b8 != 0xE7 || a8 != 0xFF {
		t.Errorf("expected background #FFF8E7FF, got #%02X%02X%02X%02X", r8, g8, b8, a8)
	}
}
