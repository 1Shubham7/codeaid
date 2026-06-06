package cmd

import (
	"bytes"
	"image"
	_ "image/png"
	"strings"

	"github.com/1shubham7/codeaid/static"
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"github.com/nfnt/resize"
)

// renderLogo converts the embedded PNG into a terminal-renderable string using
// Unicode half-block characters (▀). Each character covers 2 pixel rows: the
// foreground color is the top pixel, background is the bottom — so 1 terminal
// row = 2 image rows. This is the same technique imgcat uses internally.
func renderLogo(termWidth int) string {
	img, _, err := image.Decode(bytes.NewReader(static.LogoBytes))
	if err != nil {
		return ""
	}

	// Cap width so the logo doesn't fill the whole terminal.
	maxCols := uint(min(termWidth, 48))
	img = resize.Thumbnail(maxCols, maxCols*2, img, resize.Lanczos3)

	b := img.Bounds()
	w, h := b.Max.X, b.Max.Y
	p := termenv.ColorProfile()

	var sb strings.Builder
	for y := 0; y < h-1; y += 2 {
		for x := 0; x < w; x++ {
			top, _ := colorful.MakeColor(img.At(x, y))
			bot, _ := colorful.MakeColor(img.At(x, y+1))
			sb.WriteString(termenv.String("▀").
				Foreground(p.Color(top.Hex())).
				Background(p.Color(bot.Hex())).
				String())
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
