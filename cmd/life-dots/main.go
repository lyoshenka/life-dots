package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grin/life-dots/internal/desktop"
	"github.com/grin/life-dots/internal/model"
	"github.com/grin/life-dots/internal/render"
	"github.com/grin/life-dots/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		runServe(os.Args[1:])
		return
	}
	switch os.Args[1] {
	case "serve":
		runServe(os.Args[2:])
	case "generate":
		runGenerate(os.Args[2:])
	case "service":
		runService(os.Args[2:])
	case "schedule":
		runSchedule(os.Args[2:])
	case "help", "-h", "--help":
		printHelp()
	default:
		// allow `life-dots --addr :8080` as serve
		runServe(os.Args[1:])
	}
}

// splitFlags lets flags appear before or after the subcommand action.
func splitFlags(args []string) (flags, rest []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
			if !strings.Contains(a, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		rest = append(rest, a)
	}
	return flags, rest
}

func printHelp() {
	fmt.Print(`life-dots — single-binary life-in-dots wallpaper engine

Config is never written to disk. Studio settings live in browser localStorage;
wallpaper URLs carry flat query params (bookmark or copy that link).

Usage:
  life-dots serve [--addr :8080]
  life-dots generate [flags]
  life-dots service install|uninstall|status|print-launchd|print-systemd
  life-dots schedule install|uninstall|status

Generate flags:
  --mode life|year|year-months|quarters|monthly|weekly|custom|goal|habits
  --dob YYYY-MM-DD
  --lifespan N
  --fg #RRGGBB
  --bg #RRGGBB
  --bg-url URL
  --width N --height N
  --output FILE.png
  --set-wallpaper
`)
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address")
	// Accept but ignore legacy --data-dir so old unit files still start.
	_ = fs.String("data-dir", "", "ignored (config is URL-only)")
	flags, _ := splitFlags(args)
	_ = fs.Parse(flags)
	srv := server.New(*addr)
	fmt.Printf("life-dots listening on http://127.0.0.1%s\n", normalizeAddr(*addr))
	fmt.Printf("studio: open that URL on this Mac, then scan the QR from your iPhone\n")
	fmt.Printf("config: studio uses localStorage; wallpaper URLs use query params\n")
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}

func normalizeAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return addr
	}
	return addr
}

func runGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	mode := fs.String("mode", "", "calendar mode")
	dob := fs.String("dob", "", "birthday YYYY-MM-DD")
	lifespan := fs.Int("lifespan", 0, "lifespan years")
	fg := fs.String("fg", "", "foreground color #RRGGBB")
	bg := fs.String("bg", "", "background color #RRGGBB")
	bgURL := fs.String("bg-url", "", "background photo URL (fetched per request, not stored)")
	startDate := fs.String("start-date", "", "goal/custom start YYYY-MM-DD")
	endDate := fs.String("end-date", "", "goal/custom end YYYY-MM-DD")
	width := fs.Int("width", 0, "width px")
	height := fs.Int("height", 0, "height px")
	output := fs.String("output", "", "write PNG to path")
	setWP := fs.Bool("set-wallpaper", false, "apply to local desktop")
	_ = fs.String("data-dir", "", "ignored (config is URL-only)")
	flags, _ := splitFlags(args)
	_ = fs.Parse(flags)

	cfg := model.DefaultConfig()
	if *mode != "" {
		cfg.Mode = model.Mode(*mode)
	}
	if *dob != "" {
		cfg.Birthday = *dob
	}
	if *lifespan > 0 {
		cfg.Lifespan = *lifespan
	}
	if *fg != "" {
		cfg.Fg = *fg
	}
	if *bg != "" {
		cfg.Bg = *bg
	}
	if *bgURL != "" {
		cfg.BgType = model.BgURL
		cfg.BgURL = *bgURL
	}
	if *startDate != "" {
		cfg.StartDate = *startDate
	}
	if *endDate != "" {
		cfg.EndDate = *endDate
	}
	if (cfg.Mode == model.ModeCustom || cfg.Mode == model.ModeGoal) && cfg.StartDate == "" {
		cfg.StartDate = time.Now().Format("2006-01-02")
		cfg.EndDate = time.Now().AddDate(0, 3, 0).Format("2006-01-02")
	}
	if *width > 0 {
		cfg.Width = *width
	}
	if *height > 0 {
		cfg.Height = *height
	}

	png, err := render.RenderPNG(cfg, nil, time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
	if *output != "" {
		if err := os.WriteFile(*output, png, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("wrote", *output)
	}
	if *setWP {
		path, err := desktop.SetWallpaper(png)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wallpaper: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("desktop wallpaper set:", path)
	}
	if *output == "" && !*setWP {
		path, err := desktop.SetWallpaper(png)
		if err != nil {
			out := "wallpaper.png"
			_ = os.WriteFile(out, png, 0o644)
			fmt.Println("wrote", out)
			return
		}
		fmt.Println("wrote", path)
	}
}

func runService(args []string) {
	fs := flag.NewFlagSet("service", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address")
	_ = fs.String("data-dir", "", "ignored (legacy)")
	flags, rest := splitFlags(args)
	_ = fs.Parse(flags)
	if len(rest) == 0 {
		fmt.Println(desktop.ServiceStatus())
		return
	}
	switch rest[0] {
	case "install":
		if err := desktop.InstallService(*addr); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println("service installed")
	case "uninstall":
		if err := desktop.UninstallService(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println("service uninstalled")
	case "status":
		fmt.Println(desktop.ServiceStatus())
	case "print-launchd":
		s, err := desktop.PrintLaunchd(*addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Print(s)
	case "print-systemd":
		s, err := desktop.PrintSystemd(*addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Print(s)
	default:
		fmt.Fprintln(os.Stderr, "usage: life-dots service install|uninstall|status|print-launchd|print-systemd")
		os.Exit(1)
	}
}

func runSchedule(args []string) {
	fs := flag.NewFlagSet("schedule", flag.ExitOnError)
	_ = fs.String("data-dir", "", "ignored (legacy)")
	flags, rest := splitFlags(args)
	_ = fs.Parse(flags)
	if len(rest) == 0 {
		fmt.Println(desktop.ScheduleStatus())
		return
	}
	switch rest[0] {
	case "install":
		if err := desktop.InstallSchedule(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println("schedule installed")
	case "uninstall":
		if err := desktop.UninstallSchedule(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println("schedule uninstalled")
	case "status":
		fmt.Println(desktop.ScheduleStatus())
	default:
		fmt.Fprintln(os.Stderr, "usage: life-dots schedule install|uninstall|status")
		os.Exit(1)
	}
}
