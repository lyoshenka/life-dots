// Package fonts embeds typefaces used when rendering wallpapers (not the studio UI).
package fonts

import "embed"

// Files holds Bitcount Ink (wallpaper title/footer/axis text).
//
//go:embed BitcountInk-Regular.ttf
var Files embed.FS
