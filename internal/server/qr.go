package server

import (
	"fmt"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

// qrSVG returns an SVG QR code for the given URL (pure vector, no photo storage).
func qrSVG(content string) string {
	q, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200"><text x="10" y="20">QR error</text></svg>`
	}
	bmp := q.Bitmap()
	n := len(bmp)
	const cell = 8
	size := n * cell
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" shape-rendering="crispEdges">`, size, size, size, size)
	b.WriteString(`<rect width="100%" height="100%" fill="#fff"/>`)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if bmp[y][x] {
				fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="#000"/>`, x*cell, y*cell, cell, cell)
			}
		}
	}
	b.WriteString(`</svg>`)
	return b.String()
}
