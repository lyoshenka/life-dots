package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	LaunchdLabel = "io.grin.life-dots"
	SystemdName  = "life-dots"
)

// ExecPath returns the absolute path to the current binary.
func ExecPath() (string, error) {
	return os.Executable()
}

func logDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/life-dots"
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Logs", "life-dots")
	}
	return filepath.Join(home, ".cache", "life-dots", "logs")
}

// PrintLaunchd returns a launchd plist for the HTTP server.
func PrintLaunchd(addr string) (string, error) {
	exe, err := ExecPath()
	if err != nil {
		return "", err
	}
	if addr == "" {
		addr = ":8080"
	}
	logs := logDir()
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>serve</string>
    <string>--addr</string>
    <string>%s</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>%s/serve.out.log</string>
  <key>StandardErrorPath</key>
  <string>%s/serve.err.log</string>
</dict>
</plist>
`, LaunchdLabel, exe, addr, logs, logs), nil
}

// PrintLaunchdSchedule returns a plist that refreshes desktop wallpaper daily.
func PrintLaunchdSchedule() (string, error) {
	exe, err := ExecPath()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s.wallpaper</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>generate</string>
    <string>--set-wallpaper</string>
  </array>
  <key>StartCalendarInterval</key>
  <dict>
    <key>Hour</key>
    <integer>6</integer>
    <key>Minute</key>
    <integer>0</integer>
  </dict>
  <key>RunAtLoad</key>
  <true/>
</dict>
</plist>
`, LaunchdLabel, exe), nil
}

// PrintSystemd returns a user systemd unit for the HTTP server.
func PrintSystemd(addr string) (string, error) {
	exe, err := ExecPath()
	if err != nil {
		return "", err
	}
	if addr == "" {
		addr = ":8080"
	}
	return fmt.Sprintf(`[Unit]
Description=Life Dots wallpaper studio & image server
After=network.target

[Service]
Type=simple
ExecStart=%s serve --addr %s
Restart=on-failure
RestartSec=3

[Install]
WantedBy=default.target
`, exe, addr), nil
}

// PrintSystemdTimer returns wallpaper refresh timer + service.
func PrintSystemdTimer() (service, timer string, err error) {
	exe, err := ExecPath()
	if err != nil {
		return "", "", err
	}
	service = fmt.Sprintf(`[Unit]
Description=Life Dots desktop wallpaper refresh

[Service]
Type=oneshot
ExecStart=%s generate --set-wallpaper
`, exe)
	timer = `[Unit]
Description=Daily Life Dots wallpaper refresh

[Timer]
OnCalendar=*-*-* 06:00:00
Persistent=true

[Install]
WantedBy=timers.target
`
	return service, timer, nil
}

func launchAgentsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "LaunchAgents")
	return dir, os.MkdirAll(dir, 0o755)
}

func systemdUserDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "systemd", "user")
	return dir, os.MkdirAll(dir, 0o755)
}

// InstallService installs the HTTP daemon via launchd or systemd.
func InstallService(addr string) error {
	switch runtime.GOOS {
	case "darwin":
		dir, err := launchAgentsDir()
		if err != nil {
			return err
		}
		_ = os.MkdirAll(logDir(), 0o755)
		plist, err := PrintLaunchd(addr)
		if err != nil {
			return err
		}
		path := filepath.Join(dir, LaunchdLabel+".plist")
		if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
			return err
		}
		_ = exec.Command("launchctl", "unload", path).Run()
		out, err := exec.Command("launchctl", "load", path).CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl load: %w (%s)", err, string(out))
		}
		return nil
	case "linux":
		dir, err := systemdUserDir()
		if err != nil {
			return err
		}
		unit, err := PrintSystemd(addr)
		if err != nil {
			return err
		}
		path := filepath.Join(dir, SystemdName+".service")
		if err := os.WriteFile(path, []byte(unit), 0o644); err != nil {
			return err
		}
		cmds := [][]string{
			{"systemctl", "--user", "daemon-reload"},
			{"systemctl", "--user", "enable", "--now", SystemdName + ".service"},
		}
		for _, c := range cmds {
			out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
			if err != nil {
				return fmt.Errorf("%s: %w (%s)", strings.Join(c, " "), err, string(out))
			}
		}
		return nil
	default:
		return fmt.Errorf("service install not supported on %s", runtime.GOOS)
	}
}

// UninstallService removes the HTTP daemon.
func UninstallService() error {
	switch runtime.GOOS {
	case "darwin":
		dir, err := launchAgentsDir()
		if err != nil {
			return err
		}
		path := filepath.Join(dir, LaunchdLabel+".plist")
		_ = exec.Command("launchctl", "unload", path).Run()
		return os.Remove(path)
	case "linux":
		_ = exec.Command("systemctl", "--user", "disable", "--now", SystemdName+".service").Run()
		dir, err := systemdUserDir()
		if err != nil {
			return err
		}
		return os.Remove(filepath.Join(dir, SystemdName+".service"))
	default:
		return fmt.Errorf("unsupported OS")
	}
}

// ServiceStatus reports whether the daemon appears installed/running.
func ServiceStatus() string {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("launchctl", "list", LaunchdLabel).CombinedOutput()
		if err != nil {
			return "not loaded"
		}
		return strings.TrimSpace(string(out))
	case "linux":
		out, err := exec.Command("systemctl", "--user", "is-active", SystemdName+".service").CombinedOutput()
		if err != nil {
			return "inactive"
		}
		return strings.TrimSpace(string(out))
	default:
		return "unsupported"
	}
}

// InstallSchedule installs daily desktop wallpaper refresh.
func InstallSchedule() error {
	switch runtime.GOOS {
	case "darwin":
		dir, err := launchAgentsDir()
		if err != nil {
			return err
		}
		plist, err := PrintLaunchdSchedule()
		if err != nil {
			return err
		}
		path := filepath.Join(dir, LaunchdLabel+".wallpaper.plist")
		if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
			return err
		}
		_ = exec.Command("launchctl", "unload", path).Run()
		out, err := exec.Command("launchctl", "load", path).CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl load: %w (%s)", err, string(out))
		}
		return nil
	case "linux":
		dir, err := systemdUserDir()
		if err != nil {
			return err
		}
		svc, timer, err := PrintSystemdTimer()
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, SystemdName+"-wallpaper.service"), []byte(svc), 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, SystemdName+"-wallpaper.timer"), []byte(timer), 0o644); err != nil {
			return err
		}
		cmds := [][]string{
			{"systemctl", "--user", "daemon-reload"},
			{"systemctl", "--user", "enable", "--now", SystemdName + "-wallpaper.timer"},
		}
		for _, c := range cmds {
			out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
			if err != nil {
				return fmt.Errorf("%s: %w (%s)", strings.Join(c, " "), err, string(out))
			}
		}
		return nil
	default:
		return fmt.Errorf("schedule not supported on %s", runtime.GOOS)
	}
}

// UninstallSchedule removes the wallpaper refresh job.
func UninstallSchedule() error {
	switch runtime.GOOS {
	case "darwin":
		dir, err := launchAgentsDir()
		if err != nil {
			return err
		}
		path := filepath.Join(dir, LaunchdLabel+".wallpaper.plist")
		_ = exec.Command("launchctl", "unload", path).Run()
		return os.Remove(path)
	case "linux":
		_ = exec.Command("systemctl", "--user", "disable", "--now", SystemdName+"-wallpaper.timer").Run()
		dir, err := systemdUserDir()
		if err != nil {
			return err
		}
		_ = os.Remove(filepath.Join(dir, SystemdName+"-wallpaper.timer"))
		return os.Remove(filepath.Join(dir, SystemdName+"-wallpaper.service"))
	default:
		return fmt.Errorf("unsupported OS")
	}
}

// ScheduleStatus reports schedule state.
func ScheduleStatus() string {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("launchctl", "list", LaunchdLabel+".wallpaper").CombinedOutput()
		if err != nil {
			return "not loaded"
		}
		return strings.TrimSpace(string(out))
	case "linux":
		out, err := exec.Command("systemctl", "--user", "is-active", SystemdName+"-wallpaper.timer").CombinedOutput()
		if err != nil {
			return "inactive"
		}
		return strings.TrimSpace(string(out))
	default:
		return "unsupported"
	}
}
