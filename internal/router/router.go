package router

import (
	"github.com/go-chi/chi"
	"github.com/scrolllockdev/test-devops/internal/handlers"
	mw "github.com/scrolllockdev/test-devops/internal/middlewares"
	"github.com/scrolllockdev/test-devops/internal/repo"
)

func SetupRouter(repo repo.Repository) *chi.Mux {
	r := chi.NewRouter()
	//r.Use(middleware.Logger)
	r.Use(mw.GzipMW)
	r.Get("/", handlers.Default())
	r.Get("/value/{m_type}/{m_name}", handlers.GetMetrics(repo))
	r.Post("/update/{m_type}/{m_name}/{value}", handlers.Update(repo))
	r.Post("/update/", handlers.UpdatePost(repo))
	r.Post("/value/", handlers.GerMetricsPost(repo))

	return r
}
