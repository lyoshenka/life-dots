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
- **Appearance**: foreground / today / background colors, shapes, flat card overlay, axis labels, device presets (iPhone 13–17 Pro Max, iPad, MacBook Air/Pro, Android)
- **Photo behind dots** (fully stateless):
  - Photo URL in config — fetched per request, never stored
  - iOS Shortcut POST of a Photos album image — request-scoped only
- **Habit tracker** — log endpoints return an updated config for the client to keep in the wallpaper URL
- **CLI**: `generate`, `service` (launchd / systemd), `schedule`

## Service install

```bash
./life-dots service install --addr :8080
./life-dots schedule install   # daily desktop wallpaper at 06:00
```

## Tests

```bash
make test
```

## License

MIT
