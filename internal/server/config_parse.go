package server

import (
	"net/http"

	"github.com/grin/life-dots/internal/model"
)

func parseConfigFromRequest(r *http.Request, base model.Config) model.Config {
	cfg := base
	model.ApplyQuery(&cfg, r.URL.Query())
	return cfg
}
