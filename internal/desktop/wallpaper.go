package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// SetWallpaper writes PNG bytes to a temp file and sets it as the desktop wallpaper.
// The temp file is required by OS wallpaper APIs (macOS/Linux) — this is the rendered
// wallpaper output, not a user photo store.
func SetWallpaper(png []byte) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	outDir := filepath.Join(dir, "life-dots")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(outDir, "wallpaper.png")
	if err := os.WriteFile(path, png, 0o644); err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return path, setMacOS(path)
	case "linux":
		return path, setLinux(path)
	default:
		return path, fmt.Errorf("desktop wallpaper not supported on %s", runtime.GOOS)
	}
}

func setMacOS(path string) error {
	script := fmt.Sprintf(`tell application "System Events"
  tell every desktop
    set picture to POSIX file "%s"
  end tell
end tell`, path)
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("osascript: %w (%s)", err, string(out))
	}
	return nil
}

func setLinux(path string) error {
	// Try common backends in order
	attempts := [][]string{
		{"gsettings", "set", "org.gnome.desktop.background", "picture-uri", "file://" + path},
		{"gsettings", "set", "org.gnome.desktop.background", "picture-uri-dark", "file://" + path},
		{"feh", "--bg-fill", path},
		{"swaybg", "-i", path, "-m", "fill"},
		{"swww", "img", path},
	}
	var last error
	for _, args := range attempts {
		if _, err := exec.LookPath(args[0]); err != nil {
			continue
		}
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			last = fmt.Errorf("%s: %w (%s)", args[0], err, string(out))
			continue
		}
		return nil
	}
	if last != nil {
		return last
	}
	return fmt.Errorf("no wallpaper setter found (tried gsettings, feh, swaybg, swww)")
}
