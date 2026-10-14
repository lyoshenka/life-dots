package render

import (
	"bytes"
	"image"
	"image/color"
	"sync"

	"github.com/grin/life-dots/internal/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	bitcountOnce sync.Once
	bitcountFont *opentype.Font
	bitcountErr  error
)

func loadBitcountInk() (*opentype.Font, error) {
	bitcountOnce.Do(func() {
		data, err := fonts.Files.ReadFile("BitcountInk-Regular.ttf")
		if err != nil {
			bitcountErr = err
			return
		}
		bitcountFont, bitcountErr = opentype.Parse(data)
	})
	return bitcountFont, bitcountErr
}

func faceAt(size float64) font.Face {
	if size < 6 {
		size = 6
	}
	f, err := loadBitcountInk()
	if err != nil || f == nil {
		return nil
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	})
	if err != nil {
		return nil
	}
	return face
}

// drawText draws s centered (or left-aligned) at (x,y) baseline using Bitcount Ink.
func drawText(dst *image.RGBA, x, y int, s string, c color.RGBA, center bool, size float64) {
	if s == "" {
		return
	}
	if size <= 0 {
		size = 14
	}
	face := faceAt(size)
	if face == nil {
		return
	}
	defer func() {
		if closer, ok := face.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}()

	lines := bytes.Split([]byte(s), []byte{'\n'})
	lineH := int(size * 1.25)
	if lineH < 10 {
		lineH = 10
	}
	for li, line := range lines {
		str := string(line)
		w := font.MeasureString(face, str).Ceil()
		xx := x
		if center {
			xx = x - w/2
		}
		yy := y + li*lineH
		d := &font.Drawer{
			Dst:  dst,
			Src:  image.NewUniform(c),
			Face: face,
			Dot:  fixed.Point26_6{X: fixed.I(xx), Y: fixed.I(yy)},
		}
		d.DrawString(str)
	}
}
