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
	case "help", "-h", "--help":
		if len(os.Args) >= 3 {
			printTopicHelp(os.Args[2])
			return
		}
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

func printTopicHelp(topic string) {
	switch topic {
	case "serve":
		printServeHelp()
	case "generate":
		printGenerateHelp()
	case "service":
		printServiceHelp()
	default:
		fmt.Fprintf(os.Stderr, "unknown help topic %q\n\n", topic)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`life-dots — single-binary life-in-dots wallpaper engine

Config is never written to disk. Studio settings live in browser localStorage;
wallpaper URLs carry flat query params (bookmark or copy that link).

Usage:
  life-dots serve [--addr :8080]
  life-dots generate [flags]
  life-dots service <command> [--addr :8080]
  life-dots help [serve|generate|service]

Commands:
  serve      Run the studio + wallpaper HTTP server (default if no command)
  generate   Render a PNG locally (and optionally set desktop wallpaper)
  service    Install/uninstall the always-on HTTP server (launchd / systemd)

Examples:
  life-dots serve --addr :8080
  life-dots generate --mode year --fg '#fff' --output wallpaper.png
  life-dots service install --addr :8080

Run "life-dots help <command>" for full flags and details.
`)
}

func printServeHelp() {
	fmt.Print(`life-dots serve — studio + wallpaper HTTP server

Usage:
  life-dots serve [--addr :8080]
  life-dots [--addr :8080]          # same as serve (no subcommand)

Flags:
  --addr string   Listen address (default ":8080")

Opens a local studio UI and serves wallpaper images:
  GET  /                 Studio
  GET  /wallpaper.png    PNG with flat query-param config
  GET  /wallpaper.svg    SVG variant
  POST /wallpaper.png    Optional photo body + query config
  GET  /api/meta         Devices + habit icons
  GET  /api/qr           QR code for a wallpaper URL

Studio settings are stored in the browser only. Wallpaper links encode
settings as query params for Shortcuts / bookmarks.
`)
}

func printGenerateHelp() {
	fmt.Print(`life-dots generate — render a wallpaper PNG locally

Usage:
  life-dots generate [flags]

Flags mirror wallpaper query params (kebab-case). Only flags you pass
override defaults. Shared with /wallpaper.png via the same ApplyQuery logic.

  Calendar:
    --mode life|year|year-months|quarters|monthly|weekly|custom|goal|habits
    --birthday|--dob YYYY-MM-DD
    --lifespan N
    --year N
    --start-date YYYY-MM-DD
    --end-date|--goal-date YYYY-MM-DD
    --title string         Custom title
    --grouping string      Life grouping (none|months|decades|full|…)
    --week-start monday|sunday
    --habits JSON          Habit array JSON

  Appearance:
    --fg #RRGGBB
    --bg #RRGGBB
    --highlight-color #RRGGBB
    --fill-focus past|future
    --shape circle|rounded|square
    --shape-style filled|outlined
    --no-gap
    --highlight-today
    --show-numbers
    --show-both
    --grid-opacity 0–100

  Background:
    --bg-type match|solid|url
    --bg-url URL
    --bg-rotate 0|90|180|270
    --bg-blur N
    --bg-dim 0–80
    --bg-focal-x 0–100
    --bg-focal-y 0–100

  Card:
    --show-card
    --card-style flat
    --card-opacity 0–100
    --card-radius N
    --card-pad-h N
    --card-pad-v N

  Chrome:
    --show-axes
    --show-title
    --show-labels
    --show-branding
    --show-progress
    --mute-text
    --footer-mode none|numbers

  Layout:
    --device id             Device preset (sets width/height)
    --width N --height N
    --screen lock|home
    --widget-layout none|top|bottom|custom
    --margin N --margin-top N
    --shape-scale 50–150
    --stats-scale 50–150

  Output (CLI-only):
    --output FILE.png      Write PNG to path
    --set-wallpaper        Apply as desktop wallpaper
                           (if neither is set, tries desktop then wallpaper.png)

Examples:
  life-dots generate --mode life --dob 1995-01-01 --output out.png
  life-dots generate --mode year --fill-focus future --set-wallpaper
`)
}

func printServiceHelp() {
	fmt.Print(`life-dots service — always-on HTTP server as a user daemon

Keeps "life-dots serve" running so studio + wallpaper URLs stay available
(e.g. for iPhone Shortcuts). macOS uses launchd; Linux uses systemd --user.

Usage:
  life-dots service install [--addr :8080]
  life-dots service uninstall
  life-dots service status
  life-dots service print-launchd [--addr :8080]
  life-dots service print-systemd [--addr :8080]

Commands:
  install         Install and start the user service
  uninstall       Stop and remove the user service
  status          Show whether the service is loaded/active
  print-launchd   Print the macOS LaunchAgent plist (no install)
  print-systemd   Print the Linux user systemd unit (no install)

Flags:
  --addr string   Listen address for serve (default ":8080")

Examples:
  life-dots service install --addr :8080
  life-dots service status
`)
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.Usage = printServeHelp
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
	fs.Usage = printGenerateHelp

	// Calendar
	_ = fs.String("mode", "", "calendar mode")
	_ = fs.String("birthday", "", "birthday YYYY-MM-DD")
	_ = fs.String("dob", "", "alias for --birthday")
	_ = fs.Int("lifespan", 0, "lifespan years")
	_ = fs.Int("year", 0, "calendar year")
	_ = fs.String("start-date", "", "goal/custom start YYYY-MM-DD")
	_ = fs.String("end-date", "", "goal/custom end YYYY-MM-DD")
	_ = fs.String("goal-date", "", "alias for --end-date")
	_ = fs.String("title", "", "custom title")
	_ = fs.String("grouping", "", "life grouping")
	_ = fs.String("week-start", "", "monday|sunday")
	_ = fs.String("habits", "", "habits JSON array")

	// Appearance
	_ = fs.String("fg", "", "foreground color #RRGGBB")
	_ = fs.String("bg", "", "background color #RRGGBB")
	_ = fs.String("highlight-color", "", "today highlight color")
	_ = fs.String("fill-focus", "", "past|future")
	_ = fs.String("shape", "", "circle|rounded|square")
	_ = fs.String("shape-style", "", "filled|outlined")
	_ = fs.Bool("no-gap", false, "square with no gap")
	_ = fs.Bool("highlight-today", false, "ring around today")
	_ = fs.Bool("show-numbers", false, "show day numbers")
	_ = fs.Bool("show-both", false, "shape + number")
	_ = fs.Float64("grid-opacity", 0, "0–100")

	// Background
	_ = fs.String("bg-type", "", "match|solid|url")
	_ = fs.String("bg-url", "", "background photo URL")
	_ = fs.Int("bg-rotate", 0, "photo rotation 0|90|180|270")
	_ = fs.Float64("bg-blur", 0, "photo blur px")
	_ = fs.Float64("bg-dim", 0, "photo dim percent")
	_ = fs.Float64("bg-focal-x", 0, "focal X 0–100")
	_ = fs.Float64("bg-focal-y", 0, "focal Y 0–100")

	// Card
	_ = fs.Bool("show-card", false, "card behind dots")
	_ = fs.String("card-style", "", "flat")
	_ = fs.Float64("card-opacity", 0, "0–100")
	_ = fs.Float64("card-radius", 0, "corner radius")
	_ = fs.Float64("card-pad-h", 0, "horizontal pad")
	_ = fs.Float64("card-pad-v", 0, "vertical pad")

	// Chrome
	_ = fs.Bool("show-axes", false, "axis labels")
	_ = fs.Bool("show-title", false, "show title")
	_ = fs.Bool("show-labels", false, "show labels")
	_ = fs.Bool("show-branding", false, "show branding")
	_ = fs.Bool("show-progress", false, "show progress in footer")
	_ = fs.Bool("mute-text", false, "muted color for footer and axis labels")
	_ = fs.String("footer-mode", "", "none|numbers")

	// Layout
	_ = fs.String("device", "", "device preset id")
	_ = fs.Int("width", 0, "width px")
	_ = fs.Int("height", 0, "height px")
	_ = fs.String("screen", "", "lock|home")
	_ = fs.String("widget-layout", "", "none|top|bottom|custom")
	_ = fs.Float64("margin", 0, "side margin px")
	_ = fs.Float64("margin-top", 0, "top margin px")
	_ = fs.Float64("shape-scale", 0, "50–150")
	_ = fs.Float64("stats-scale", 0, "50–150")

	// CLI-only
	output := fs.String("output", "", "write PNG to path")
	setWP := fs.Bool("set-wallpaper", false, "apply to local desktop")
	_ = fs.String("data-dir", "", "ignored (config is URL-only)")

	flags, _ := splitFlags(args)
	_ = fs.Parse(flags)

	cfg := model.DefaultConfig()
	q := map[string][]string{}
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "output", "set-wallpaper", "data-dir":
			return
		}
		key := strings.ReplaceAll(f.Name, "-", "_")
		q[key] = []string{f.Value.String()}
	})
	model.ApplyQuery(&cfg, q)

	if (cfg.Mode == model.ModeCustom || cfg.Mode == model.ModeGoal) && cfg.StartDate == "" {
		cfg.StartDate = time.Now().Format("2006-01-02")
		cfg.EndDate = time.Now().AddDate(0, 3, 0).Format("2006-01-02")
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
	fs.Usage = printServiceHelp
	addr := fs.String("addr", ":8080", "listen address")
	_ = fs.String("data-dir", "", "ignored (legacy)")
	flags, rest := splitFlags(args)
	_ = fs.Parse(flags)
	if len(rest) == 0 {
		printServiceHelp()
		return
	}
	switch rest[0] {
	case "help", "-h", "--help":
		printServiceHelp()
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
		fmt.Fprintf(os.Stderr, "unknown service command %q\n\n", rest[0])
		printServiceHelp()
		os.Exit(1)
	}
}

