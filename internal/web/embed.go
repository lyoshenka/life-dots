package web

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/grin/life-dots/internal/fonts"
	"github.com/grin/life-dots/internal/model"
)

//go:embed templates/* static/*
var assets embed.FS

var studioTmpl *template.Template
var pwaTmpl *template.Template

func init() {
	studioTmpl = template.Must(template.New("studio.html").Funcs(template.FuncMap{
		"json": func(v any) template.JS {
			// filled at render via companion — keep simple
			return ""
		},
	}).ParseFS(assets, "templates/studio.html"))
	pwaTmpl = template.Must(template.ParseFS(assets, "templates/pwa.html"))
}

// StudioData is passed to the studio template.
type StudioData struct {
	Config     model.Config
	Host       string
	Devices    []model.Device
	Icons      []model.ShapePath
	ConfigJSON template.JS
}

// RenderStudio writes the main studio page.
func RenderStudio(w io.Writer, cfg model.Config, host string) error {
	data := StudioData{
		Config:  cfg,
		Host:    host,
		Devices: model.Devices,
		Icons:   model.HabitIcons,
	}
	return studioTmpl.ExecuteTemplate(w, "studio.html", data)
}

// RenderPWA writes the iPhone PWA page.
func RenderPWA(w io.Writer, cfg model.Config, host string) error {
	data := StudioData{Config: cfg, Host: host}
	return pwaTmpl.ExecuteTemplate(w, "pwa.html", data)
}

// ServeStatic serves embedded CSS/JS and local font files.
func ServeStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	var data []byte
	var err error
	if strings.HasPrefix(path, "fonts/") {
		data, err = fonts.Files.ReadFile(strings.TrimPrefix(path, "fonts/"))
	} else {
		data, err = fs.ReadFile(assets, "static/"+path)
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch {
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case strings.HasSuffix(path, ".svg"):
		w.Header().Set("Content-Type", "image/svg+xml")
	case strings.HasSuffix(path, ".ttf"):
		w.Header().Set("Content-Type", "font/ttf")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	_, _ = w.Write(data)
}
