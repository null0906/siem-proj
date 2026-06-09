package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	r.Get("/api/health", h.Health)
	r.Get("/api/deployment", h.GetDeployment)
	r.Get("/api/dashboard", h.GetDashboard)
	r.Get("/api/pipeline", h.GetPipeline)
	r.Get("/api/ingest/status", h.GetPipeline)
	r.Get("/api/findings/summary", h.GetFindingsSummary)
	r.Get("/api/identity/summary", h.GetIdentitySummary)
	r.Get("/api/findings", h.ListFindings)
	r.Get("/api/findings/{id}", h.GetFinding)
	r.Get("/api/sources", h.ListSources)
	r.Get("/api/system", h.GetSystem)

	return r
}
