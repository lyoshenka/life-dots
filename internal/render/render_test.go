package render_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

func TestRotateImage(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 10, 20))
	src.Set(0, 0, color.RGBA{255, 0, 0, 255})
	r90 := render.RotateImage(src, 90)
	if r90.Bounds().Dx() != 20 || r90.Bounds().Dy() != 10 {
		t.Fatalf("90: %v", r90.Bounds())
	}
	r180 := render.RotateImage(src, 180)
	if r180.Bounds().Dx() != 10 || r180.Bounds().Dy() != 20 {
		t.Fatalf("180: %v", r180.Bounds())
	}
	r270 := render.RotateImage(src, 270)
	if r270.Bounds().Dx() != 20 || r270.Bounds().Dy() != 10 {
		t.Fatalf("270: %v", r270.Bounds())
	}
	if render.RotateImage(src, 0).Bounds() != src.Bounds() {
		t.Fatal("0 should keep size")
	}
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

func TestFetchPhotoURLCache(t *testing.T) {
	render.ClearPhotoURLCache()
	t.Cleanup(render.ClearPhotoURLCache)

	var hits atomic.Int32
	tiny := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var pngBuf bytes.Buffer
	_ = png.Encode(&pngBuf, tiny)
	pngBytes := pngBuf.Bytes()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytes)
	}))
	t.Cleanup(ts.Close)

	img1, err := render.FetchPhotoURL(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	img2, err := render.FetchPhotoURL(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("expected 1 fetch, got %d", hits.Load())
	}
	if img1 == nil || img2 == nil {
		t.Fatal("nil image")
	}

	// Fill cache past capacity; first URL should be evicted.
	for i := 0; i < render.PhotoURLCacheCap; i++ {
		u := fmt.Sprintf("%s?n=%d", ts.URL, i)
		if _, err := render.FetchPhotoURL(u); err != nil {
			t.Fatal(err)
		}
	}
	if render.PhotoURLCacheLen() != render.PhotoURLCacheCap {
		t.Fatalf("cache len=%d", render.PhotoURLCacheLen())
	}
	before := hits.Load()
	if _, err := render.FetchPhotoURL(ts.URL); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != before+1 {
		t.Fatalf("evicted URL should refetch: hits %d→%d", before, hits.Load())
	}
}

func TestFetchPhotoURLCacheExpiry(t *testing.T) {
	render.ClearPhotoURLCache()
	t.Cleanup(func() {
		render.ClearPhotoURLCache()
		render.SetPhotoCacheNow(nil)
	})

	var hits atomic.Int32
	tiny := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var pngBuf bytes.Buffer
	_ = png.Encode(&pngBuf, tiny)
	pngBytes := pngBuf.Bytes()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write(pngBytes)
	}))
	t.Cleanup(ts.Close)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := base
	render.SetPhotoCacheNow(func() time.Time { return now })

	if _, err := render.FetchPhotoURL(ts.URL); err != nil {
		t.Fatal(err)
	}
	now = base.Add(13 * 24 * time.Hour)
	if _, err := render.FetchPhotoURL(ts.URL); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Fatalf("should still be cached before TTL, hits=%d", hits.Load())
	}
	now = base.Add(14*24*time.Hour + time.Second)
	if _, err := render.FetchPhotoURL(ts.URL); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("expired entry should refetch, hits=%d", hits.Load())
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
