package server

import (
	"encoding/json"
	"image"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/grin/life-dots/internal/desktop"
	"github.com/grin/life-dots/internal/model"
	"github.com/grin/life-dots/internal/render"
	"github.com/grin/life-dots/internal/web"
)

// Server is the life-dots HTTP service. Studio state is client localStorage; wallpapers use flat URL query params.
type Server struct {
	Addr string
	mux  *http.ServeMux
}

// New creates a server.
func New(addr string) *Server {
	s := &Server{Addr: addr, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleStudio)
	s.mux.HandleFunc("/pwa", s.handlePWA)
	s.mux.HandleFunc("/static/", web.ServeStatic)

	s.mux.HandleFunc("/wallpaper.png", s.handleWallpaperPNG)
	s.mux.HandleFunc("/wallpaper.svg", s.handleWallpaperSVG)
	s.mux.HandleFunc("/p/", s.handleProfileWallpaper)

	s.mux.HandleFunc("/life", s.modeWallpaper(model.ModeLife))
	s.mux.HandleFunc("/days", s.modeWallpaper(model.ModeYear))
	s.mux.HandleFunc("/months", s.modeWallpaper(model.ModeYearMonths))
	s.mux.HandleFunc("/quarters", s.modeWallpaper(model.ModeQuarters))
	s.mux.HandleFunc("/goal", s.modeWallpaper(model.ModeGoal))
	s.mux.HandleFunc("/habits", s.modeWallpaper(model.ModeHabits))

	s.mux.HandleFunc("/api/habits/log", s.handleHabitLog)
	s.mux.HandleFunc("/api/habits/toggle", s.handleHabitToggle)
	s.mux.HandleFunc("/api/qr", s.handleQR)
	s.mux.HandleFunc("/api/desktop/apply", s.handleDesktopApply)
	s.mux.HandleFunc("/api/meta", s.handleMeta)
	s.mux.HandleFunc("/api/preview.svg", s.handleWallpaperSVG)
}

// Handler returns the HTTP handler (for ListenAndServe and tests).
func (s *Server) Handler() http.Handler {
	return s.mux
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.Addr, s.mux)
}

func (s *Server) handleStudio(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	host := publicHost(r, s.Addr)
	_ = web.RenderStudio(w, model.DefaultConfig(), host)
}

func (s *Server) handlePWA(w http.ResponseWriter, r *http.Request) {
	host := publicHost(r, s.Addr)
	_ = web.RenderPWA(w, model.DefaultConfig(), host)
}

func (s *Server) resolveConfig(r *http.Request) model.Config {
	return parseConfigFromRequest(r, model.DefaultConfig())
}

func (s *Server) readPhoto(r *http.Request) (image.Image, error) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return nil, nil
	}
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/") {
		if err := r.ParseMultipartForm(render.MaxPhotoBytes); err != nil {
			return nil, err
		}
		file, _, err := r.FormFile("photo")
		if err != nil {
			file, _, err = r.FormFile("file")
		}
		if err != nil {
			file, _, err = r.FormFile("image")
		}
		if err != nil {
			return nil, nil
		}
		defer file.Close()
		return render.DecodePhoto(file)
	}
	if strings.HasPrefix(ct, "image/") || ct == "application/octet-stream" || ct == "" {
		defer r.Body.Close()
		return render.DecodePhoto(io.LimitReader(r.Body, render.MaxPhotoBytes+1))
	}
	return nil, nil
}

func (s *Server) writePNG(w http.ResponseWriter, r *http.Request, cfg model.Config) {
	photo, _ := s.readPhoto(r)
	png, err := render.RenderPNG(cfg, photo, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}

func (s *Server) handleWallpaperPNG(w http.ResponseWriter, r *http.Request) {
	s.writePNG(w, r, s.resolveConfig(r))
}

func (s *Server) handleWallpaperSVG(w http.ResponseWriter, r *http.Request) {
	cfg := s.resolveConfig(r)
	svg, err := render.RenderSVG(cfg, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(svg))
}

func (s *Server) handleProfileWallpaper(w http.ResponseWriter, r *http.Request) {
	// /p/... is a wallpaper alias — all config comes from query params (no server-side profiles).
	path := strings.TrimPrefix(r.URL.Path, "/p/")
	path = strings.TrimSuffix(path, ".png")
	path = strings.TrimSuffix(path, ".svg")
	parts := strings.Split(path, "/")
	cfg := s.resolveConfig(r)

	// /p/*/habit/{habitId}/log.png — mark habit done for this request only, then render
	if len(parts) >= 4 && parts[1] == "habit" && strings.HasPrefix(parts[3], "log") {
		habitID := parts[2]
		date := time.Now().Format("2006-01-02")
		cfg = applyHabit(cfg, habitID, date, true)
		s.writePNG(w, r, cfg)
		return
	}

	if strings.HasSuffix(r.URL.Path, ".svg") {
		svg, err := render.RenderSVG(cfg, time.Now())
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(svg))
		return
	}
	s.writePNG(w, r, cfg)
}

func (s *Server) modeWallpaper(mode model.Mode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := s.resolveConfig(r)
		cfg.Mode = mode
		if mode == model.ModeGoal {
			cfg.Mode = model.ModeCustom
		}
		s.writePNG(w, r, cfg)
	}
}

func applyHabit(cfg model.Config, habitID, date string, done bool) model.Config {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	found := false
	for i := range cfg.Habits {
		if cfg.Habits[i].ID == habitID {
			if cfg.Habits[i].Completions == nil {
				cfg.Habits[i].Completions = map[string]bool{}
			}
			cfg.Habits[i].Completions[date] = done
			found = true
			break
		}
	}
	if !found && len(cfg.Habits) > 0 {
		if cfg.Habits[0].Completions == nil {
			cfg.Habits[0].Completions = map[string]bool{}
		}
		cfg.Habits[0].Completions[date] = done
	}
	return cfg
}

func (s *Server) handleHabitLog(w http.ResponseWriter, r *http.Request) {
	habitID := r.URL.Query().Get("habit")
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	done := r.URL.Query().Get("done") != "0"
	if r.Method == http.MethodPost {
		var body struct {
			HabitID string `json:"habitId"`
			Date    string `json:"date"`
			Done    *bool  `json:"done"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.HabitID != "" {
			habitID = body.HabitID
		}
		if body.Date != "" {
			date = body.Date
		}
		if body.Done != nil {
			done = *body.Done
		}
	}
	cfg := applyHabit(s.resolveConfig(r), habitID, date, done)
	if r.URL.Query().Get("wallpaper") == "1" || strings.Contains(r.Header.Get("Accept"), "image/png") {
		s.writePNG(w, r, cfg)
		return
	}
	// Return updated config so the client can put it back in the wallpaper URL.
	writeJSON(w, map[string]any{
		"config": cfg,
	})
}

func (s *Server) handleHabitToggle(w http.ResponseWriter, r *http.Request) {
	s.handleHabitLog(w, r)
}

func (s *Server) handleQR(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		url = publicHost(r, s.Addr) + "/wallpaper.png"
	}
	svg := qrSVG(url)
	w.Header().Set("Content-Type", "image/svg+xml")
	_, _ = w.Write([]byte(svg))
}

func (s *Server) handleDesktopApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	cfg := s.resolveConfig(r)
	png, err := render.RenderPNG(cfg, nil, time.Now())
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	path, err := desktop.SetWallpaper(png)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]string{"path": path, "status": "ok"})
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"devices":    model.Devices,
		"habitIcons": model.HabitIcons,
		"host":       publicHost(r, s.Addr),
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func publicHost(r *http.Request, listenAddr string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xf := r.Header.Get("X-Forwarded-Proto"); xf != "" {
		scheme = xf
	}
	host := r.Host
	if host == "" {
		host = listenAddr
		if strings.HasPrefix(host, ":") {
			if ip := localIP(); ip != "" {
				host = ip + host
			} else {
				host = "127.0.0.1" + host
			}
		}
	}
	return scheme + "://" + host
}

func localIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip != nil {
				return ip.String()
			}
		}
	}
	return ""
}
