# Life Dots

Single-binary Go + CSS wallpaper engine that shows your life in dots on iPhone (via Shortcuts), Mac, or Linux.

## Quick start

```bash
make dev
```

Open `http://127.0.0.1:8080` on your Mac, configure your calendar, scan the QR with your iPhone, and set up a daily Shortcuts automation that fetches the wallpaper URL.

**Studio settings** are saved in the browser’s `localStorage` (the address bar stays clean). **Wallpaper URLs** carry settings as normal query params — copy that link for Shortcuts or bookmarks.

## Features

- **Modes**: life, year days, year months, quarters, monthly, weekly, goal/custom, habits
- **Appearance**: primary / other / today / background colors, dot spacing, past vs future highlight, shapes, Bitcount Ink on wallpaper text, flat card overlay, axis labels, device presets (iPhone 13–17 Pro Max, iPad, MacBook Air/Pro, Android)
- **Photo behind dots**:
  - Photo URL in config — fetched on demand, kept in an in-memory LRU cache (up to 20 URLs, 14-day TTL)
  - iOS Shortcut POST of a Photos album image — request-scoped only (not cached)
- **Habit tracker** — log endpoints return an updated config for the client to keep in the wallpaper URL

## Build

```bash
make build         # native → dist/life-dots
make build-mac     # macOS → dist/life-dots-mac
make build-linux   # Linux → dist/life-dots-linux
```

## CLI

```bash
./dist/life-dots help                 # overview
./dist/life-dots help serve
./dist/life-dots help generate
./dist/life-dots help service
```

### serve

Runs the studio + wallpaper HTTP server (also the default if you omit a command).

```bash
./dist/life-dots serve --addr :8080
```

### generate

Renders a PNG locally. Flags mirror wallpaper query params (kebab-case); only flags you pass override defaults.

```bash
./dist/life-dots generate --mode year --fg '#ffffff' --output wallpaper.png
./dist/life-dots generate --mode life --dob 1995-01-01 --set-wallpaper
```

### service

Installs an always-on user daemon so `serve` stays up (launchd on macOS, systemd --user on Linux) — useful for iPhone Shortcuts hitting wallpaper URLs.

```bash
./dist/life-dots service install --addr :8080
./dist/life-dots service status
./dist/life-dots service uninstall
./dist/life-dots service print-launchd --addr :8080   # print plist only
./dist/life-dots service print-systemd --addr :8080   # print unit only
```

## Tests

```bash
make test
```

## License

MIT
