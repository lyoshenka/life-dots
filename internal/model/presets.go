package model

import (
	"fmt"
	"strings"
)

// Device is a screen resolution preset.
type Device struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Group  string `json:"group"`
}

// Devices is the full catalog from TheLifeCalendar / OneDot / WeeklyDots.
var Devices = []Device{
	// iPhone
	{ID: "iphone-se", Name: "iPhone SE", Width: 750, Height: 1334, Group: "iPhone"},
	{ID: "iphone-13-mini", Name: "iPhone 13 mini", Width: 1080, Height: 2340, Group: "iPhone"},
	{ID: "iphone-13", Name: "iPhone 13 / 13 Pro / 14", Width: 1170, Height: 2532, Group: "iPhone"},
	{ID: "iphone-14-pro", Name: "iPhone 14 Pro", Width: 1179, Height: 2556, Group: "iPhone"},
	{ID: "iphone-13-pro-max", Name: "iPhone 13 Pro Max / 14 Plus", Width: 1284, Height: 2778, Group: "iPhone"},
	{ID: "iphone-15", Name: "iPhone 15 / 15 Pro / 16", Width: 1179, Height: 2556, Group: "iPhone"},
	{ID: "iphone-15-plus", Name: "iPhone 15 Plus / 15 Pro Max / 16 Plus", Width: 1290, Height: 2796, Group: "iPhone"},
	{ID: "iphone-16-pro", Name: "iPhone 16 Pro", Width: 1206, Height: 2622, Group: "iPhone"},
	{ID: "iphone-16-pro-max", Name: "iPhone 16 Pro Max", Width: 1320, Height: 2868, Group: "iPhone"},
	{ID: "iphone-17", Name: "iPhone 17", Width: 1179, Height: 2556, Group: "iPhone"},
	{ID: "iphone-17-pro", Name: "iPhone 17 Pro", Width: 1206, Height: 2622, Group: "iPhone"},
	{ID: "iphone-17-pro-max", Name: "iPhone 17 Pro Max", Width: 1320, Height: 2868, Group: "iPhone"},
	// iPad
	{ID: "ipad", Name: "iPad", Width: 1620, Height: 2160, Group: "iPad"},
	{ID: "ipad-pro-11", Name: "iPad Pro 11\"", Width: 1668, Height: 2388, Group: "iPad"},
	{ID: "ipad-pro-13", Name: "iPad Pro 13\"", Width: 2064, Height: 2752, Group: "iPad"},
	// Mac
	{ID: "macbook-air-13", Name: "MacBook Air 13\"", Width: 2560, Height: 1664, Group: "Mac"},
	{ID: "macbook-air-15", Name: "MacBook Air 15\"", Width: 2880, Height: 1864, Group: "Mac"},
	{ID: "macbook-pro-14", Name: "MacBook Pro 14\"", Width: 3024, Height: 1964, Group: "Mac"},
	{ID: "macbook-pro-16", Name: "MacBook Pro 16\"", Width: 3456, Height: 2234, Group: "Mac"},
	{ID: "imac-4k", Name: "iMac 4K", Width: 4096, Height: 2304, Group: "Mac"},
	{ID: "imac-5k", Name: "iMac 5K", Width: 5120, Height: 2880, Group: "Mac"},
	// Android
	{ID: "samsung-s24", Name: "Samsung Galaxy S24", Width: 1080, Height: 2340, Group: "Android"},
	{ID: "samsung-s24-ultra", Name: "Samsung Galaxy S24 Ultra", Width: 1440, Height: 3120, Group: "Android"},
	{ID: "pixel-9", Name: "Google Pixel 9", Width: 1080, Height: 2424, Group: "Android"},
	{ID: "pixel-9-pro", Name: "Google Pixel 9 Pro", Width: 1280, Height: 2856, Group: "Android"},
	{ID: "xiaomi-14", Name: "Xiaomi 14", Width: 1200, Height: 2670, Group: "Android"},
	{ID: "oneplus-12", Name: "OnePlus 12", Width: 1440, Height: 3168, Group: "Android"},
	{ID: "sony-xperia", Name: "Sony Xperia", Width: 1080, Height: 2520, Group: "Android"},
	{ID: "nothing-phone", Name: "Nothing Phone", Width: 1080, Height: 2412, Group: "Android"},
	{ID: "fairphone", Name: "Fairphone", Width: 1080, Height: 2400, Group: "Android"},
	{ID: "generic-hd", Name: "Generic HD", Width: 720, Height: 1280, Group: "Generic"},
	{ID: "generic-fhd", Name: "Generic FHD", Width: 1080, Height: 1920, Group: "Generic"},
	{ID: "generic-qhd", Name: "Generic QHD", Width: 1440, Height: 2560, Group: "Generic"},
	{ID: "generic-4k", Name: "Generic 4K", Width: 2160, Height: 3840, Group: "Generic"},
}

// DeviceByID returns a device preset or nil.
func DeviceByID(id string) *Device {
	for i := range Devices {
		if Devices[i].ID == id {
			return &Devices[i]
		}
	}
	return nil
}

// ResolveColors builds a palette from fg/bg (with optional custom overrides).
func ResolveColors(cfg Config) Colors {
	fg := normalizeHex(cfg.Fg)
	bg := normalizeHex(cfg.Bg)
	if fg == "" {
		fg = "#FFFFFF"
	}
	if bg == "" {
		bg = "#000000"
	}
	muted := mixHex(bg, fg, 0.22)
	past, future := fg, muted
	if cfg.FillFocus == FillFuture {
		past, future = muted, fg
	}
	c := Colors{
		Background: bg,
		Past:       past,
		Current:    fg,
		Future:     future,
		Text:       fg,
		TextMuted:  mixHex(bg, fg, 0.5),
		Accent:     fg,
	}
	if cfg.CustomColors != nil {
		if cfg.CustomColors.Background != "" {
			c.Background = cfg.CustomColors.Background
		}
		if cfg.CustomColors.Past != "" {
			c.Past = cfg.CustomColors.Past
		}
		if cfg.CustomColors.Current != "" {
			c.Current = cfg.CustomColors.Current
		}
		if cfg.CustomColors.Future != "" {
			c.Future = cfg.CustomColors.Future
		}
		if cfg.CustomColors.Text != "" {
			c.Text = cfg.CustomColors.Text
		}
		if cfg.CustomColors.TextMuted != "" {
			c.TextMuted = cfg.CustomColors.TextMuted
		}
		if cfg.CustomColors.Accent != "" {
			c.Accent = cfg.CustomColors.Accent
		}
	}
	if cfg.HighlightColor != "" {
		c.Current = cfg.HighlightColor
	}
	return c
}

func normalizeHex(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if s[0] != '#' {
		s = "#" + s
	}
	if len(s) == 4 { // #RGB
		return string([]byte{'#', s[1], s[1], s[2], s[2], s[3], s[3]})
	}
	if len(s) >= 7 {
		return strings.ToUpper(s[:7])
	}
	return strings.ToUpper(s)
}

func mixHex(a, b string, t float64) string {
	ar, ag, ab := parseRGB(a)
	br, bg, bb := parseRGB(b)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	r := int(float64(ar)*(1-t) + float64(br)*t)
	g := int(float64(ag)*(1-t) + float64(bg)*t)
	bl := int(float64(ab)*(1-t) + float64(bb)*t)
	return fmt.Sprintf("#%02X%02X%02X", r, g, bl)
}

func parseRGB(hex string) (r, g, b int) {
	h := normalizeHex(hex)
	if len(h) < 7 {
		return 0, 0, 0
	}
	fmt.Sscanf(h, "#%02x%02x%02x", &r, &g, &b)
	return
}
