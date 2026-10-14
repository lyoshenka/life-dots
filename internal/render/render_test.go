package render_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/grin/life-dots/internal/model"
	"github.com/grin/life-dots/internal/render"
)

func TestCoverCropAndBlur(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 100, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 100; x++ {
			src.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	out := render.CoverCrop(src, 50, 80, 50, 50)
	if out.Bounds().Dx() != 50 || out.Bounds().Dy() != 80 {
		t.Fatal(out.Bounds())
	}
	blurred := render.BoxBlur(out, 2)
	if blurred.Bounds() != out.Bounds() {
		t.Fatal("blur size")
	}
	render.ApplyDim(blurred, 40, color.RGBA{})
}

func TestDecodePhoto(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	got, err := render.DecodePhoto(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if got.Bounds().Dx() != 8 {
		t.Fatal(got.Bounds())
	}
}

func TestRenderNumbersAndFooter(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Width, cfg.Height = 200, 400
	cfg.Mode = model.ModeLife
	cfg.ShowNumbers = true
	cfg.ShowBoth = true
	cfg.FooterMode = model.FooterNumbers
	data, err := render.RenderPNG(cfg, nil, time.Date(2026, 10, 14, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
}
