package server

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/scrolllockdev/test-devops/internal/model"
	"github.com/scrolllockdev/test-devops/internal/server/config"
	"github.com/scrolllockdev/test-devops/internal/server/database"
	"github.com/scrolllockdev/test-devops/internal/server/handlers"
	"github.com/scrolllockdev/test-devops/internal/server/middlewares"
	"github.com/scrolllockdev/test-devops/internal/server/storage"
)

func InitLogger() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
}

type Server struct {
	r        *chi.Mux
	server   *http.Server
	Storage  storage.Storage
	Config   config.Config
	Database *sql.DB
}

func (server *Server) Init() *Server {

	InitLogger()

	cfg := config.Config{}
	if err := cfg.ReadConfig(); err != nil {
		log.Fatalln(err)
	}

	server.Config = cfg
	server.r = chi.NewRouter()
	server.server = &http.Server{
		Addr:    server.Config.ServerAddress,
		Handler: server.r,
	}
	server.Storage = storage.Storage{
		Storage: make([]model.Metric, 0),
	}

	if server.Config.DatabaseDsn != "" {
		db, err := sql.Open("postgres", server.Config.DatabaseDsn)
		if err != nil {
			log.Fatalln(err)
		} else {
			server.Database = db
		}
	} else {
		log.Infoln("there is no connection to database")
		server.Database = nil
	}

	return server
}

func (server *Server) Run(ctx context.Context) {
	if server.Config.StoragePath != "" {
		if server.Config.Restore {
			err := server.Storage.RestoreFromFile(server.Config.StoragePath)
			if err != nil {
				log.Errorln(err)
			} else {
				log.Infoln("metrics successfuly restored")
			}
		}
		if server.Config.StoreInterval > 0*time.Second {
			storeTicker := time.NewTicker(server.Config.StoreInterval)
			go func(ctx context.Context) {
				for {
					select {
					case <-storeTicker.C:
						if err := server.Storage.StoreToFile(server.Config.StoragePath); err != nil {
							log.Errorln(err)
						}
					case <-ctx.Done():
						storeTicker.Stop()
						return
					}
				}
			}(ctx)
		}
	}

	// creation storage table
	if err := database.CreateStorageTable(ctx, server.Database, "storage"); err != nil {
		log.Errorln(err)
	} else {
		log.Infoln("table successfully created")
	}

	server.r.Use(middlewares.Gzip)
	server.r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricFromAddress(&server.Storage, server.Config))
	server.r.Get("/value/{type}/{name}", handlers.GetMetricFromAddress(&server.Storage))
	server.r.Get("/", handlers.AllMetrics(&server.Storage))
	server.r.Post("/value/", handlers.GetMetricValueFromBody(&server.Storage, server.Config.Key))

	hashes := server.r.Group(nil)
	hashes.Use(middlewares.EqualHashes(server.Config.Key))
	hashes.Post("/update/", handlers.UpdateMetricFromBody(&server.Storage, server.Config, server.Database))

	database := server.r.Group(nil)
	database.Get("/ping", handlers.PingDB(server.Config.DatabaseDsn))
	database.Post("/updates/", handlers.UpdateMetrics(&server.Storage, server.Database))

	go server.server.ListenAndServe()

}

func (server *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
	if err := server.Storage.StoreToFile(server.Config.StoragePath); err != nil {
		log.Errorln(err)
	}
	if server.Database != nil {
		server.Database.Close()
	}
}
