package export

import (
	"image"
	"image/color"
	"strings"

	"github.com/fogleman/gg"
)

// RenderOptions holds configuration for napkin image rendering.
type RenderOptions struct {
	Content  string
	Width    int
	Height   int
	FontSize float64
	FontPath string
}

// RenderNapkin renders a note as a napkin-style image.
func RenderNapkin(opts RenderOptions) (image.Image, error) {
	dc := gg.NewContext(opts.Width, opts.Height)

	// Draw napkin background (#FFF8E7)
	dc.SetColor(color.RGBA{R: 0xFF, G: 0xF8, B: 0xE7, A: 0xFF})
	dc.Clear()

	// Draw subtle horizontal lines every 30px
	lineColor := color.RGBA{R: 0xD0, G: 0xC8, B: 0xB8, A: 0x40}
	dc.SetColor(lineColor)
	dc.SetLineWidth(0.5)
	for y := 30.0; y < float64(opts.Height); y += 30 {
		dc.DrawLine(0, y, float64(opts.Width), y)
		dc.Stroke()
	}

	// Load font if specified, otherwise use default
	if opts.FontPath != "" {
		if err := dc.LoadFontFace(opts.FontPath, opts.FontSize); err != nil {
			// Fall back to default if custom font fails
			dc.LoadFontFace("", opts.FontSize)
		}
	}

	// Draw text with word wrapping
	if opts.Content != "" {
		dc.SetColor(color.RGBA{R: 0x2D, G: 0x2D, B: 0x2D, A: 0xFF})

		padding := 40.0
		lineHeight := opts.FontSize * 1.6
		maxWidth := float64(opts.Width) - (padding * 2)

		lines := wordWrap(dc, opts.Content, maxWidth)
		y := padding + opts.FontSize
		for _, line := range lines {
			dc.DrawStringAnchored(line, padding, y, 0, 0.5)
			y += lineHeight
			if y > float64(opts.Height)-padding {
				break
			}
		}
	}

	return dc.Image(), nil
}

// wordWrap breaks text into lines that fit within maxWidth.
func wordWrap(dc *gg.Context, text string, maxWidth float64) []string {
	var lines []string
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			testLine := currentLine + " " + word
			w, _ := dc.MeasureString(testLine)
			if w <= maxWidth {
				currentLine = testLine
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
		lines = append(lines, currentLine)
	}

	return lines
}
