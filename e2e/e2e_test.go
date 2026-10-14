package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"net/url"
	"strconv"

	"github.com/grin/life-dots/internal/calendar"
	"github.com/grin/life-dots/internal/desktop"
	"github.com/grin/life-dots/internal/model"
	"github.com/grin/life-dots/internal/render"
	"github.com/grin/life-dots/internal/server"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func testNow() time.Time {
	return time.Date(2026, 10, 14, 12, 0, 0, 0, time.Local)
}

func baseCfg() model.Config {
	cfg := model.DefaultConfig()
	cfg.Birthday = "1995-01-01"
	cfg.Lifespan = 90
	cfg.Year = 2026
	cfg.Width = 200
	cfg.Height = 400
	cfg.StartDate = "2026-01-01"
	cfg.EndDate = "2026-06-01"
	cfg.Habits = []model.Habit{
		{ID: "h1", Name: "Move", Icon: "dumbbell", Completions: map[string]bool{"2026-10-14": true}},
		{ID: "h2", Name: "Read", Icon: "book-open", Completions: map[string]bool{}},
	}
	return cfg
}

func queryFromCfg(cfg model.Config) string {
	q := url.Values{}
	q.Set("mode", string(cfg.Mode))
	q.Set("birthday", cfg.Birthday)
	q.Set("lifespan", strconv.Itoa(cfg.Lifespan))
	q.Set("year", strconv.Itoa(cfg.Year))
	q.Set("start_date", cfg.StartDate)
	q.Set("end_date", cfg.EndDate)
	q.Set("title", cfg.CustomTitle)
	q.Set("fg", cfg.Fg)
	q.Set("bg", cfg.Bg)
	q.Set("width", strconv.Itoa(cfg.Width))
	q.Set("height", strconv.Itoa(cfg.Height))
	q.Set("device", cfg.DeviceID)
	if len(cfg.Habits) > 0 {
		raw, _ := json.Marshal(cfg.Habits)
		q.Set("habits", string(raw))
	}
	return q.Encode()
}

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 32, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 8), uint8(y * 4), 80, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func assertPNG(t *testing.T, data []byte, wantW, wantH int) {
	t.Helper()
	if len(data) < 100 {
		t.Fatalf("png too small: %d bytes", len(data))
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("png decode: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != wantW || b.Dy() != wantH {
		t.Fatalf("png size got %dx%d want %dx%d", b.Dx(), b.Dy(), wantW, wantH)
	}
}

func assertSVG(t *testing.T, s string) {
	t.Helper()
	if !strings.Contains(s, "<svg") || !strings.Contains(s, "</svg>") {
		t.Fatalf("not svg: %q", s[:min(80, len(s))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- Calendar engine (all modes) ---

func TestE2E_CalendarAllModes(t *testing.T) {
	now := testNow()
	modes := []model.Mode{
		model.ModeLife,
		model.ModeYear,
		model.ModeYearMonths,
		model.ModeQuarters,
		model.ModeMonthly,
		model.ModeWeekly,
		model.ModeCustom,
		model.ModeGoal,
		model.ModeHabits,
	}
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			cfg := baseCfg()
			cfg.Mode = mode
			g, err := calendar.Build(cfg, now)
			if err != nil {
				t.Fatal(err)
			}
			if g.Columns < 1 || g.Rows < 1 {
				t.Fatalf("bad grid %dx%d", g.Columns, g.Rows)
			}
			if len(g.Cells) != g.Columns*g.Rows {
				t.Fatalf("cells=%d want %d", len(g.Cells), g.Columns*g.Rows)
			}
			if g.Total < 1 {
				t.Fatalf("total=%d", g.Total)
			}
		})
	}
}

func TestE2E_CalendarGroupingsAndWeekStart(t *testing.T) {
	now := testNow()
	cfg := baseCfg()
	cfg.Mode = model.ModeLife
	for _, g := range []model.Grouping{model.GroupingNone, model.GroupingMonths, model.GroupingDecades, model.GroupingFull} {
		cfg.Grouping = g
		if _, err := calendar.Build(cfg, now); err != nil {
			t.Fatalf("grouping %s: %v", g, err)
		}
	}
	cfg.WeekStart = model.WeekStartSunday
	cfg.Mode = model.ModeMonthly
	if _, err := calendar.Build(cfg, now); err != nil {
		t.Fatal(err)
	}
	cfg.Mode = model.ModeWeekly
	cfg.Grouping = model.GroupingRows
	if _, err := calendar.Build(cfg, now); err != nil {
		t.Fatal(err)
	}
}

func TestE2E_CalendarSpecialDatesAndHeatmap(t *testing.T) {
	cfg := baseCfg()
	cfg.Mode = model.ModeYear
	cfg.SpecialDates = []model.SpecialDate{{StartDate: "2026-10-01", EndDate: "2026-10-20", Color: "#ff00ff"}}
	cfg.DailyCounts = map[string]int{"2026-10-14": 5, "2026-10-13": 2}
	g, err := calendar.Build(cfg, testNow())
	if err != nil {
		t.Fatal(err)
	}
	foundSpecial, foundHeat := false, false
	for _, c := range g.Cells {
		if c.State == calendar.CellSpecial || c.Color == "#ff00ff" {
			foundSpecial = true
		}
		if c.Opacity > 0 && c.Opacity < 1 {
			foundHeat = true
		}
	}
	if !foundSpecial {
		t.Fatal("expected special date cell")
	}
	if !foundHeat {
		t.Fatal("expected heatmap opacity")
	}
}

// --- Rendering ---

func TestE2E_RenderPNGAllModes(t *testing.T) {
	now := testNow()
	modes := []model.Mode{
		model.ModeLife, model.ModeYear, model.ModeYearMonths, model.ModeQuarters,
		model.ModeMonthly, model.ModeWeekly,
		model.ModeCustom, model.ModeHabits,
	}
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			cfg := baseCfg()
			cfg.Mode = mode
			cfg.ShowCard = true
			cfg.ShowAxes = true
			cfg.ShowProgress = true
			cfg.FooterMode = model.FooterNumbers
			cfg.HighlightToday = true
			data, err := render.RenderPNG(cfg, nil, now)
			if err != nil {
				t.Fatal(err)
			}
			assertPNG(t, data, cfg.Width, cfg.Height)
		})
	}
}

func TestE2E_RenderSVGAllModes(t *testing.T) {
	now := testNow()
	for _, mode := range []model.Mode{model.ModeLife, model.ModeYear, model.ModeCustom, model.ModeHabits} {
		cfg := baseCfg()
		cfg.Mode = mode
		svg, err := render.RenderSVG(cfg, now)
		if err != nil {
			t.Fatal(err)
		}
		assertSVG(t, svg)
	}
}

func TestE2E_RenderShapesAndThemes(t *testing.T) {
	now := testNow()
	cfg := baseCfg()
	cfg.Mode = model.ModeLife
	for _, shape := range []model.DotShape{model.ShapeCircle, model.ShapeRounded, model.ShapeSquare} {
		cfg.DotShape = shape
		cfg.NoGap = shape == model.ShapeSquare
		cfg.ShapeStyle = model.StyleOutlined
		cfg.OutlinedFillPast = true
		if _, err := render.RenderPNG(cfg, nil, now); err != nil {
			t.Fatalf("shape %s: %v", shape, err)
		}
	}
	for _, pair := range [][2]string{{"#FFFFFF", "#000000"}, {"#38BDF8", "#0A1628"}, {"#1A1A1A", "#FFFFFF"}} {
		cfg.Fg, cfg.Bg = pair[0], pair[1]
		cfg.ShapeStyle = model.StyleFilled
		if _, err := render.RenderPNG(cfg, nil, now); err != nil {
			t.Fatalf("colors %s/%s: %v", pair[0], pair[1], err)
		}
	}
}

func TestE2E_RenderPhotoPassThrough(t *testing.T) {
	cfg := baseCfg()
	cfg.Mode = model.ModeLife
	cfg.BgDim = 40
	cfg.BgBlur = 2
	cfg.ShowCard = true
	cfg.CardStyle = model.CardFlat
	cfg.GridOpacity = 80

	photo, _, err := image.Decode(bytes.NewReader(tinyPNG(t)))
	if err != nil {
		t.Fatal(err)
	}
	data, err := render.RenderPNG(cfg, photo, testNow())
	if err != nil {
		t.Fatal(err)
	}
	assertPNG(t, data, cfg.Width, cfg.Height)
}

func TestE2E_RenderPhotoURL(t *testing.T) {
	// Local httptest serving a PNG as the "remote" photo — no retention after request.
	photoBytes := tinyPNG(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(photoBytes)
	}))
	t.Cleanup(ts.Close)

	cfg := baseCfg()
	cfg.Mode = model.ModeYear
	cfg.BgType = model.BgURL
	cfg.BgURL = ts.URL
	cfg.BgFocalX = 30
	cfg.BgFocalY = 70
	cfg.BgBlur = 1
	cfg.BgDim = 25
	data, err := render.RenderPNG(cfg, nil, testNow())
	if err != nil {
		t.Fatal(err)
	}
	assertPNG(t, data, cfg.Width, cfg.Height)
}

func TestE2E_RenderMatchBg(t *testing.T) {
	cfg := baseCfg()
	cfg.Mode = model.ModeLife
	cfg.BgType = model.BgMatch
	if _, err := render.RenderPNG(cfg, nil, testNow()); err != nil {
		t.Fatal(err)
	}
}

func TestE2E_DevicePresetsRender(t *testing.T) {
	cfg := baseCfg()
	cfg.Mode = model.ModeLife
	// Use smaller sizes for speed but validate catalog IDs resolve.
	for _, id := range []string{"iphone-se", "iphone-16-pro", "iphone-16-pro-max", "generic-fhd"} {
		d := model.DeviceByID(id)
		if d == nil {
			t.Fatalf("missing device %s", id)
		}
		cfg.DeviceID = d.ID
		cfg.Width = 120
		cfg.Height = 240
		if _, err := render.RenderPNG(cfg, nil, testNow()); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
	}
}

// --- Desktop service unit printers ---

func TestE2E_DesktopPrintUnits(t *testing.T) {
	plist, err := desktop.PrintLaunchd(":9090")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(plist, "io.grin.life-dots") || !strings.Contains(plist, "serve") {
		t.Fatalf("bad plist: %s", plist[:min(200, len(plist))])
	}
	if strings.Contains(plist, "--data-dir") {
		t.Fatal("launchd should not use --data-dir")
	}
	unit, err := desktop.PrintSystemd(":9090")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(unit, "[Service]") || !strings.Contains(unit, "serve --addr") {
		t.Fatalf("bad unit: %s", unit)
	}
	if strings.Contains(unit, "--data-dir") {
		t.Fatal("systemd should not use --data-dir")
	}
}

// --- HTTP server e2e ---

func newTestServer(t *testing.T) *server.Server {
	t.Helper()
	return server.New(":0")
}

func TestE2E_HTTPStudioAndStatic(t *testing.T) {
	srv := newTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 || !bytes.Contains(body, []byte("Life Dots")) {
		t.Fatalf("studio: %d %s", res.StatusCode, body[:min(100, len(body))])
	}

	res, err = http.Get(ts.URL + "/pwa")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ = io.ReadAll(res.Body)
	if res.StatusCode != 200 || !bytes.Contains(body, []byte("apple-mobile-web-app-capable")) {
		t.Fatal("pwa missing")
	}

	res, err = http.Get(ts.URL + "/static/styles.css")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ = io.ReadAll(res.Body)
	if res.StatusCode != 200 || !bytes.Contains(body, []byte("--accent")) {
		t.Fatal("css missing")
	}

	res, err = http.Get(ts.URL + "/static/studio.js")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatal("js missing")
	}
}

func TestE2E_HTTPWallpaperEndpoints(t *testing.T) {
	srv := newTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	paths := []string{
		"/wallpaper.png?mode=life&birthday=1995-01-01&width=180&height=360",
		"/wallpaper.svg?mode=life&width=180&height=360",
		"/life?birthday=1995-01-01&width=180&height=360",
		"/days?width=180&height=360",
		"/months?width=180&height=360",
		"/quarters?width=180&height=360",
		"/goal?goal=Race&start_date=2026-01-01&goal_date=2026-12-31&width=180&height=360",
		"/habits?width=180&height=360",
		"/api/preview.svg?mode=year",
	}
	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			res, err := http.Get(ts.URL + p)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			if res.StatusCode != 200 {
				t.Fatalf("status %d: %s", res.StatusCode, body)
			}
			ct := res.Header.Get("Content-Type")
			if strings.Contains(p, ".svg") || strings.HasSuffix(strings.Split(p, "?")[0], ".svg") || strings.Contains(p, "preview.svg") || strings.Contains(p, "wallpaper.svg") {
				assertSVG(t, string(body))
			} else {
				if !strings.Contains(ct, "image/png") {
					t.Fatalf("ct=%s", ct)
				}
				assertPNG(t, body, 180, 360)
			}
			if res.Header.Get("Cache-Control") != "no-store" && strings.Contains(ct, "image/") {
				t.Fatalf("expected no-store for %s", p)
			}
		})
	}
}

func TestE2E_HTTPPostPhotoWallpaper(t *testing.T) {
	srv := newTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	photo := tinyPNG(t)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/wallpaper.png?mode=life&width=160&height=320&bg_dim=30", bytes.NewReader(photo))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "image/png")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatalf("%d %s", res.StatusCode, body)
	}
	assertPNG(t, body, 160, 320)

	// multipart form
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("photo", "bg.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write(photo)
	_ = w.Close()
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/life?width=160&height=320", &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ = io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatalf("multipart %d", res.StatusCode)
	}
	assertPNG(t, body, 160, 320)
}

func TestE2E_HTTPConfigViaQueryParam(t *testing.T) {
	srv := newTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	cfg := baseCfg()
	cfg.CustomTitle = "FROM QUERY"
	cfg.Width = 150
	cfg.Height = 300
	q := queryFromCfg(cfg)

	res, err := http.Get(ts.URL + "/wallpaper.png?" + q)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatalf("%d %s", res.StatusCode, body)
	}
	assertPNG(t, body, 150, 300)

	res, err = http.Get(ts.URL + "/p/any.png?" + q)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ = io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatal(res.StatusCode)
	}
	assertPNG(t, body, 150, 300)

	res, err = http.Get(ts.URL + "/api/config")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 404 {
		t.Fatalf("expected 404 for /api/config, got %d", res.StatusCode)
	}
}

func TestE2E_HTTPHabitLogAndWallpaper(t *testing.T) {
	srv := newTestServer(t)
	cfg := baseCfg()
	cfg.Mode = model.ModeHabits
	cfg.Width = 140
	cfg.Height = 280
	q := queryFromCfg(cfg)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	res, err := http.Post(ts.URL+"/api/habits/log?habit=h1&done=1&wallpaper=1&"+q, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatalf("%d %s", res.StatusCode, body)
	}
	assertPNG(t, body, 140, 280)

	res, err = http.Get(ts.URL + "/p/x/habit/h2/log.png?" + q)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ = io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatal(res.StatusCode)
	}
	assertPNG(t, body, 140, 280)

	res, err = http.Post(ts.URL+"/api/habits/log?habit=h2&done=1&"+q, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["config"] == nil {
		t.Fatal("expected updated config")
	}
}

func TestE2E_HTTPMetaAndQR(t *testing.T) {
	srv := newTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/api/meta")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var meta map[string]json.RawMessage
	if err := json.NewDecoder(res.Body).Decode(&meta); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"devices", "habitIcons"} {
		if _, ok := meta[key]; !ok {
			t.Fatalf("meta missing %s", key)
		}
	}
	for _, gone := range []string{"quotes", "shapes", "bgPresets"} {
		if _, ok := meta[gone]; ok {
			t.Fatalf("meta should not include removed %s", gone)
		}
	}

	res, err = http.Get(ts.URL + "/api/qr?url=http://example.com/wallpaper.png")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatal(res.StatusCode)
	}
	assertSVG(t, string(body))
}

func TestE2E_HTTPQueryOverrides(t *testing.T) {
	srv := newTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	photoTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(tinyPNG(t))
	}))
	t.Cleanup(photoTS.Close)

	url := fmt.Sprintf("%s/wallpaper.png?mode=custom&start_date=2026-01-01&end_date=2026-02-01&fg=%%2338BDF8&bg=%%230A1628&shape=rounded&grouping=weeks&title=Hello&bg_url=%s&bg_blur=1&bg_dim=20&width=170&height=340&show_card=1",
		ts.URL, photoTS.URL)
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatalf("%d %s", res.StatusCode, body)
	}
	assertPNG(t, body, 170, 340)
}

// --- CLI e2e ---

func buildCLI(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "life-dots")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/life-dots")
	cmd.Dir = filepath.Join("..")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func TestE2E_CLIGenerate(t *testing.T) {
	bin := buildCLI(t)
	outPNG := filepath.Join(t.TempDir(), "out.png")

	cmd := exec.Command(bin, "generate",
		"--mode", "life",
		"--dob", "1995-01-01",
		"--lifespan", "80",
		"--fg", "#FFFFFF",
		"--bg", "#000000",
		"--bg-type", "solid",
		"--fill-focus", "future",
		"--shape", "rounded",
		"--week-start", "sunday",
		"--highlight-color", "#FF6B4A",
		"--grid-opacity", "90",
		"--show-title",
		"--footer-mode", "numbers",
		"--width", "160",
		"--height", "320",
		"--output", outPNG,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate: %v\n%s", err, out)
	}
	data, err := os.ReadFile(outPNG)
	if err != nil {
		t.Fatal(err)
	}
	assertPNG(t, data, 160, 320)

	cmd = exec.Command(bin, "generate",
		"--mode", "custom",
		"--start-date", "2026-01-01",
		"--end-date", "2026-03-01",
		"--width", "160",
		"--height", "320",
		"--output", filepath.Join(t.TempDir(), "goal.png"),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate custom: %v\n%s", err, out)
	}
}

func TestE2E_CLIServicePrint(t *testing.T) {
	bin := buildCLI(t)

	cmd := exec.Command(bin, "service", "print-launchd", "--addr", ":7777")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !bytes.Contains(out, []byte("io.grin.life-dots")) {
		t.Fatal(string(out))
	}
	if bytes.Contains(out, []byte("--data-dir")) {
		t.Fatal("launchd should not use data-dir")
	}

	cmd = exec.Command(bin, "service", "print-systemd")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !bytes.Contains(out, []byte("[Unit]")) || !bytes.Contains(out, []byte("serve --addr")) {
		t.Fatalf("%s", out)
	}

	cmd = exec.Command(bin, "service", "status")
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Logf("service status: %v", err)
	}
}

func TestE2E_CLIHelp(t *testing.T) {
	bin := buildCLI(t)
	cmd := exec.Command(bin, "help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"serve", "generate", "service", "URL", "help <command>"} {
		if !bytes.Contains(out, []byte(s)) {
			t.Fatalf("help missing %s", s)
		}
	}
	if bytes.Contains(out, []byte("config show")) {
		t.Fatal("help should not advertise config persistence")
	}
	if bytes.Contains(out, []byte("schedule")) {
		t.Fatal("help should not mention removed schedule command")
	}

	topics := map[string][]string{
		"serve":    {"--addr", "wallpaper.png", "Studio"},
		"generate": {"--fill-focus", "--bg-type", "--bg-rotate", "--set-wallpaper", "--output"},
		"service":  {"install", "print-launchd", "print-systemd", "launchd"},
	}
	for topic, needles := range topics {
		cmd = exec.Command(bin, "help", topic)
		out, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("help %s: %v\n%s", topic, err, out)
		}
		for _, s := range needles {
			if !bytes.Contains(out, []byte(s)) {
				t.Fatalf("help %s missing %q\n%s", topic, s, out)
			}
		}
	}
}

func TestE2E_ModelCatalogs(t *testing.T) {
	if len(model.Devices) < 20 {
		t.Fatalf("devices=%d", len(model.Devices))
	}
	if len(model.HabitIcons) < 20 {
		t.Fatalf("icons=%d", len(model.HabitIcons))
	}
	if model.DeviceByID("iphone-16-pro") == nil {
		t.Fatal("device by id")
	}
	if model.ShapeByID("dumbbell") == nil {
		t.Fatal("habit icon")
	}
	cfg := baseCfg()
	cfg.Fg = "#38BDF8"
	cfg.Bg = "#0A1628"
	c := model.ResolveColors(cfg)
	if c.Background != "#0A1628" || c.Past != "#38BDF8" {
		t.Fatalf("resolve colors: %+v", c)
	}
	cfg.FillFocus = model.FillFuture
	c = model.ResolveColors(cfg)
	if c.Future != "#38BDF8" || c.Past == "#38BDF8" {
		t.Fatalf("fill focus future: %+v", c)
	}
}
