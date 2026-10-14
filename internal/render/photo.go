package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/grin/life-dots/internal/model"
)

// MaxPhotoBytes limits a single request-scoped photo decode.
const MaxPhotoBytes = 25 << 20 // 25 MiB

// PhotoURLCacheCap is the max number of decoded background photos kept in memory.
const PhotoURLCacheCap = 20

// PhotoURLCacheTTL is how long a cached photo stays valid after fetch.
const PhotoURLCacheTTL = 14 * 24 * time.Hour

type photoCacheEntry struct {
	img      image.Image
	storedAt time.Time
}

var (
	photoCacheMu    sync.Mutex
	photoCache      = map[string]photoCacheEntry{}
	photoCacheOrder []string // LRU: oldest → newest
	photoCacheNow   = time.Now
)

// FetchPhotoURL downloads and decodes a photo, caching up to PhotoURLCacheCap URLs
// in memory (LRU). Entries expire after PhotoURLCacheTTL.
func FetchPhotoURL(url string) (image.Image, error) {
	if url == "" {
		return nil, fmt.Errorf("empty bg url")
	}
	if img, ok := photoCacheGet(url); ok {
		return img, nil
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bg url status %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, MaxPhotoBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxPhotoBytes {
		return nil, fmt.Errorf("photo too large")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	photoCachePut(url, img)
	return img, nil
}

func photoCacheGet(url string) (image.Image, bool) {
	photoCacheMu.Lock()
	defer photoCacheMu.Unlock()
	photoCacheEvictExpiredLocked()
	e, ok := photoCache[url]
	if !ok {
		return nil, false
	}
	if photoCacheNow().Sub(e.storedAt) >= PhotoURLCacheTTL {
		photoCacheDeleteLocked(url)
		return nil, false
	}
	// Move to most-recently used.
	for i, u := range photoCacheOrder {
		if u == url {
			photoCacheOrder = append(append(photoCacheOrder[:i], photoCacheOrder[i+1:]...), url)
			break
		}
	}
	return e.img, true
}

func photoCachePut(url string, img image.Image) {
	photoCacheMu.Lock()
	defer photoCacheMu.Unlock()
	photoCacheEvictExpiredLocked()
	now := photoCacheNow()
	if _, ok := photoCache[url]; ok {
		photoCache[url] = photoCacheEntry{img: img, storedAt: now}
		for i, u := range photoCacheOrder {
			if u == url {
				photoCacheOrder = append(append(photoCacheOrder[:i], photoCacheOrder[i+1:]...), url)
				return
			}
		}
		photoCacheOrder = append(photoCacheOrder, url)
		return
	}
	for len(photoCacheOrder) >= PhotoURLCacheCap {
		evict := photoCacheOrder[0]
		photoCacheOrder = photoCacheOrder[1:]
		delete(photoCache, evict)
	}
	photoCache[url] = photoCacheEntry{img: img, storedAt: now}
	photoCacheOrder = append(photoCacheOrder, url)
}

func photoCacheEvictExpiredLocked() {
	now := photoCacheNow()
	kept := photoCacheOrder[:0]
	for _, u := range photoCacheOrder {
		e, ok := photoCache[u]
		if !ok || now.Sub(e.storedAt) >= PhotoURLCacheTTL {
			delete(photoCache, u)
			continue
		}
		kept = append(kept, u)
	}
	photoCacheOrder = kept
}

func photoCacheDeleteLocked(url string) {
	delete(photoCache, url)
	for i, u := range photoCacheOrder {
		if u == url {
			photoCacheOrder = append(photoCacheOrder[:i], photoCacheOrder[i+1:]...)
			return
		}
	}
}

// SetPhotoCacheNow overrides the clock used for cache TTL (tests). Pass nil to restore.
func SetPhotoCacheNow(fn func() time.Time) {
	photoCacheMu.Lock()
	defer photoCacheMu.Unlock()
	if fn == nil {
		photoCacheNow = time.Now
		return
	}
	photoCacheNow = fn
}

// ClearPhotoURLCache drops all cached background photos (tests / memory pressure).
func ClearPhotoURLCache() {
	photoCacheMu.Lock()
	defer photoCacheMu.Unlock()
	photoCache = map[string]photoCacheEntry{}
	photoCacheOrder = nil
}

// PhotoURLCacheLen returns the number of cached URLs (tests).
func PhotoURLCacheLen() int {
	photoCacheMu.Lock()
	defer photoCacheMu.Unlock()
	photoCacheEvictExpiredLocked()
	return len(photoCache)
}

// DecodePhoto reads image bytes from a POST body (request-scoped).
func DecodePhoto(r io.Reader) (image.Image, error) {
	limited := io.LimitReader(r, MaxPhotoBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxPhotoBytes {
		return nil, fmt.Errorf("photo too large")
	}
	// Try multipart not needed here — caller extracts body.
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// CoverCrop scales src to cover dstW×dstH with focal point (0–100).
func CoverCrop(src image.Image, dstW, dstH int, focalX, focalY float64) *image.RGBA {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	if sw == 0 || sh == 0 || dstW == 0 || dstH == 0 {
		return image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	}
	scale := math.Max(float64(dstW)/float64(sw), float64(dstH)/float64(sh))
	nw := int(math.Ceil(float64(sw) * scale))
	nh := int(math.Ceil(float64(sh) * scale))
	scaled := resizeNearest(src, nw, nh)

	if focalX < 0 {
		focalX = 50
	}
	if focalY < 0 {
		focalY = 50
	}
	maxX := nw - dstW
	maxY := nh - dstH
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	ox := int(float64(maxX) * focalX / 100)
	oy := int(float64(maxY) * focalY / 100)

	out := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	draw.Draw(out, out.Bounds(), scaled, image.Pt(ox, oy), draw.Src)
	return out
}

func resizeNearest(src image.Image, w, h int) *image.RGBA {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		sy := sb.Min.Y + y*sh/h
		for x := 0; x < w; x++ {
			sx := sb.Min.X + x*sw/w
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

// BoxBlur applies a separable box blur (approximation of Gaussian). radius in pixels.
func BoxBlur(src *image.RGBA, radius int) *image.RGBA {
	if radius <= 0 {
		return src
	}
	if radius > 40 {
		radius = 40
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	tmp := image.NewRGBA(b)
	out := image.NewRGBA(b)
	// Horizontal
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var rSum, gSum, bSum, aSum, n int
			for k := -radius; k <= radius; k++ {
				xx := x + k
				if xx < 0 {
					xx = 0
				}
				if xx >= w {
					xx = w - 1
				}
				c := src.RGBAAt(xx, y)
				rSum += int(c.R)
				gSum += int(c.G)
				bSum += int(c.B)
				aSum += int(c.A)
				n++
			}
			tmp.SetRGBA(x, y, color.RGBA{uint8(rSum / n), uint8(gSum / n), uint8(bSum / n), uint8(aSum / n)})
		}
	}
	// Vertical
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var rSum, gSum, bSum, aSum, n int
			for k := -radius; k <= radius; k++ {
				yy := y + k
				if yy < 0 {
					yy = 0
				}
				if yy >= h {
					yy = h - 1
				}
				c := tmp.RGBAAt(x, yy)
				rSum += int(c.R)
				gSum += int(c.G)
				bSum += int(c.B)
				aSum += int(c.A)
				n++
			}
			out.SetRGBA(x, y, color.RGBA{uint8(rSum / n), uint8(gSum / n), uint8(bSum / n), uint8(aSum / n)})
		}
	}
	return out
}

// ApplyDim overlays a dark scrim (dimPercent 0–80).
func ApplyDim(dst *image.RGBA, dimPercent float64, tint color.RGBA) {
	if dimPercent <= 0 {
		return
	}
	if dimPercent > 80 {
		dimPercent = 80
	}
	alpha := uint8(dimPercent / 100 * 255)
	if tint.A == 0 {
		tint = color.RGBA{0, 0, 0, alpha}
	} else {
		tint.A = alpha
	}
	b := dst.Bounds()
	overlay := image.NewUniform(tint)
	draw.Draw(dst, b, overlay, image.Point{}, draw.Over)
}

// DrawRoundedRect fills a rounded rectangle with optional border.
func DrawRoundedRect(dst *image.RGBA, r image.Rectangle, radius float64, fill color.RGBA, borderW float64, border color.RGBA) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			d := sdfRoundRect(float64(x)+0.5, float64(y)+0.5, float64(r.Min.X), float64(r.Min.Y), float64(r.Dx()), float64(r.Dy()), radius)
			if d < -borderW {
				dst.SetRGBA(x, y, alphaOver(dst.RGBAAt(x, y), fill))
			} else if d < 0 && borderW > 0 {
				dst.SetRGBA(x, y, alphaOver(dst.RGBAAt(x, y), border))
			} else if d < 1 {
				// soft edge
				a := uint8(math.Max(0, math.Min(255, (1-d)*float64(fill.A))))
				c := fill
				c.A = a
				dst.SetRGBA(x, y, alphaOver(dst.RGBAAt(x, y), c))
			}
		}
	}
}

func sdfRoundRect(px, py, x, y, w, h, radius float64) float64 {
	cx := x + w/2
	cy := y + h/2
	dx := math.Abs(px-cx) - w/2 + radius
	dy := math.Abs(py-cy) - h/2 + radius
	ax := math.Max(dx, 0)
	ay := math.Max(dy, 0)
	return math.Min(math.Max(dx, dy), 0) + math.Hypot(ax, ay) - radius
}

func alphaOver(under, over color.RGBA) color.RGBA {
	oa := float64(over.A) / 255
	ua := float64(under.A) / 255
	outA := oa + ua*(1-oa)
	if outA == 0 {
		return color.RGBA{}
	}
	r := (float64(over.R)*oa + float64(under.R)*ua*(1-oa)) / outA
	g := (float64(over.G)*oa + float64(under.G)*ua*(1-oa)) / outA
	b := (float64(over.B)*oa + float64(under.B)*ua*(1-oa)) / outA
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(outA * 255)}
}

// PrepareBackground builds the wallpaper background from config + optional request photo.
// photo may be nil; if cfg.BgType==url, fetches cfg.BgURL for this call only.
func PrepareBackground(cfg model.Config, w, h int, photo image.Image) (*image.RGBA, error) {
	colors := model.ResolveColors(cfg)
	bg := image.NewRGBA(image.Rect(0, 0, w, h))
	base := parseHex(colors.Background)
	draw.Draw(bg, bg.Bounds(), &image.Uniform{base}, image.Point{}, draw.Src)

	var src image.Image
	var err error
	switch cfg.BgType {
	case model.BgURL:
		if photo != nil {
			src = photo
		} else if cfg.BgURL != "" {
			src, err = FetchPhotoURL(cfg.BgURL)
			if err != nil {
				return bg, nil // fall back to solid
			}
		}
	case model.BgMatch:
		// dark tint of accent
		ac := parseHex(colors.Accent)
		ac.R = ac.R / 6
		ac.G = ac.G / 6
		ac.B = ac.B / 6
		draw.Draw(bg, bg.Bounds(), &image.Uniform{ac}, image.Point{}, draw.Src)
	default:
		// solid already drawn
	}
	if photo != nil && cfg.BgType != model.BgURL {
		// POST photo pass-through always composites regardless of bgType
		src = photo
	}

	if src != nil {
		src = RotateImage(src, cfg.BgRotate)
		cropped := CoverCrop(src, w, h, cfg.BgFocalX, cfg.BgFocalY)
		if cfg.BgBlur > 0 {
			cropped = BoxBlur(cropped, int(cfg.BgBlur))
		}
		draw.Draw(bg, bg.Bounds(), cropped, image.Point{}, draw.Src)
		if cfg.BgDim > 0 {
			ApplyDim(bg, cfg.BgDim, color.RGBA{0, 0, 0, 0})
		}
	}
	return bg, nil
}

// RotateImage rotates src clockwise by deg (0, 90, 180, or 270).
func RotateImage(src image.Image, deg int) image.Image {
	if src == nil {
		return nil
	}
	deg = ((deg % 360) + 360) % 360
	// Snap to nearest supported quarter-turn.
	switch {
	case deg > 315 || deg <= 45:
		return src
	case deg <= 135:
		deg = 90
	case deg <= 225:
		deg = 180
	default:
		deg = 270
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	switch deg {
	case 90:
		out := image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				out.Set(h-1-y, x, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return out
	case 180:
		out := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				out.Set(w-1-x, h-1-y, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return out
	case 270:
		out := image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				out.Set(y, w-1-x, src.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return out
	default:
		return src
	}
}

func parseHex(s string) color.RGBA {
	if s == "" {
		return color.RGBA{0, 0, 0, 255}
	}
	if s[0] == '#' {
		s = s[1:]
	}
	var r, g, b int
	var a int = 255
	switch len(s) {
	case 6:
		fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
	case 8:
		fmt.Sscanf(s, "%02x%02x%02x%02x", &r, &g, &b, &a)
	case 3:
		fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
		r *= 17
		g *= 17
		b *= 17
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

// DrawGlassCard paints a flat translucent rounded card behind the dots.
func DrawGlassCard(bg *image.RGBA, rect image.Rectangle, cfg model.Config) {
	if !cfg.ShowCard {
		return
	}
	radius := cfg.CardRadius
	if radius < 0 {
		radius = 0
	}
	fill := parseHex(cfg.CardColor)
	if cfg.CardColor == "" {
		fill = color.RGBA{255, 255, 255, 255}
	}
	op := cfg.CardOpacity
	if op < 0 {
		op = 0
	}
	if op > 100 {
		op = 100
	}
	fill.A = uint8(op / 100 * 255)

	border := parseHex(cfg.CardBorderColor)
	if cfg.CardBorderColor == "" {
		border = color.RGBA{255, 255, 255, 80}
	}
	DrawRoundedRect(bg, rect, radius, fill, cfg.CardBorderW, border)
}
