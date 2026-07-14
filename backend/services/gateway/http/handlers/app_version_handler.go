package handlers

import "net/http"

// AppVersionConfig is one app's (driver/rider) current version gate,
// loaded from env vars at startup (see gateway/cmd/server/main.go) — no
// database table; an MVP admin wanting to change this edits env vars and
// redeploys. Flutter does the actual version comparison/decision itself
// (see plan's Startup Flow phase) — this handler just reports the numbers.
type AppVersionConfig struct {
	MinimumVersion string
	LatestVersion  string
	ForceUpdate    bool
}

// AppVersionHandler serves the App Version startup check — public, no auth
// (it must work before a client has ever logged in).
type AppVersionHandler struct {
	configs map[string]AppVersionConfig
}

// NewAppVersionHandler constructs an AppVersionHandler. configs is keyed by
// app name ("driver", "rider").
func NewAppVersionHandler(configs map[string]AppVersionConfig) *AppVersionHandler {
	return &AppVersionHandler{configs: configs}
}

// GetVersion handles GET /api/v1/app/version?app=driver|rider.
func (h *AppVersionHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	app := r.URL.Query().Get("app")
	cfg, ok := h.configs[app]
	if !ok {
		writeBadRequest(w, `app must be "driver" or "rider"`)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"minimum_version": cfg.MinimumVersion,
		"latest_version":  cfg.LatestVersion,
		"force_update":    cfg.ForceUpdate,
	})
}
