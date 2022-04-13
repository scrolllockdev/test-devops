package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/scrolllockdev/test-devops/internal/server/config"
	mw "github.com/scrolllockdev/test-devops/internal/server/middlewares"
	"github.com/scrolllockdev/test-devops/internal/server/model"
	"github.com/scrolllockdev/test-devops/internal/server/storage"
)

type Server struct {
	r             *chi.Mux
	server        *http.Server
	address       string
	storeInterval time.Duration
	dbPath        string
	restore       bool
	storage       storage.Storage
}

func (s *Server) Init(cfg config.Config) *Server {

	s.r = chi.NewRouter()
	s.address = cfg.ServerAddress
	s.storeInterval = cfg.StoreInterval
	s.dbPath = cfg.StorePath
	s.restore = cfg.Restore

	s.server = &http.Server{
		Addr:    s.address,
		Handler: s.r,
	}

	s.storage = storage.Storage{
		Storage: make([]model.Metric, 0),
	}

	return s
}

func (s *Server) Run(ctx context.Context) {

	if s.dbPath != "" {

		if s.restore {
			err := s.restoreFromFile(&s.storage)
			if err != nil {
				fmt.Println(err)
			}
		}

		if s.storeInterval > 0*time.Second {
			storeTicker := time.NewTicker(s.storeInterval)

			go func(ctx context.Context) {
				for {
					select {
					case <-storeTicker.C:
						if err := s.storeToFile(); err != nil {
							fmt.Println(err)
						}
					case <-ctx.Done():
						storeTicker.Stop()
						return
					}
				}
			}(ctx)
		}
	}
	s.r.Use(mw.GzipMW)
	s.r.Post("/update/{type}/{name}/{value}", s.updateMetric())
	s.r.Get("/value/{type}/{name}", s.currentMetric())
	s.r.Get("/", s.allMetrics())
	s.r.Post("/update/", s.updateMetricFromBody())
	s.r.Post("/value/", s.getMetricValueFromBody())
	go s.server.ListenAndServe()

}

func (s *Server) restoreFromFile(storage *storage.Storage) error {
	pwd, _ := os.Getwd()
	tmpDirEx, err := storage.DirExist(path.Join(pwd, "tmp"))
	if err != nil {
		return err
	}
	if !tmpDirEx {
		fmt.Println("nothing to restore")
		return nil
	} else {
		fmt.Println("metrics restored")
	}
	file, err := os.OpenFile(path.Join(pwd, s.dbPath), os.O_RDONLY, 0755)
	var buf bytes.Buffer
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				buf.Write(line)
				break
			} else {
				return err
			}
		}
		buf.Write(line)
	}
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf.Bytes(), &s.storage)

	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func (s *Server) storeToFile() error {
	pwd, _ := os.Getwd()
	tmpDirEx, err := s.storage.DirExist(path.Join(pwd, "tmp"))
	if err != nil {
		return err
	}
	if !tmpDirEx {
		if err := s.storage.CreateDir(); err != nil {
			return err
		}
	}

	data, _ := json.MarshalIndent(s.storage, "", "  ")

	data = append(data, '\n')

	file, err := os.OpenFile(path.Join(pwd, s.dbPath), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func (s *Server) updateMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stat := strings.Split(r.URL.String(), "/")
		if err, statusCode := s.storage.UpdateMetric(stat); err != nil {
			http.Error(w, err.Error(), statusCode)
		} else {
			if s.storeInterval == 0 && s.dbPath != "" {
				if err := s.storeToFile(); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
			w.WriteHeader(statusCode)

			return
		}
	}
}

func (s *Server) currentMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stat := strings.Split(r.URL.String(), "/")
		if err, statusCode, value := s.storage.GetMetric(stat); err != nil {
			http.Error(w, err.Error(), statusCode)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(statusCode)
			w.Write(value)
			return
		}
	}
}

//func Default() http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Add("Accept-Encoding", "gzip")
//		w.Header().Add("Content-Type", "text/html")
//		w.WriteHeader(http.StatusOK)
//		io.WriteString(w, "<html><body>"+strings.Repeat("Hello, it's metrics server<br>", 20)+"</body></html>")
//	}
//}
func (s *Server) allMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Accept-Encoding", "gzip")
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		for item := range s.storage.Storage {
			io.WriteString(w, s.storage.Storage[item].ID+"\n")
		}
		// res := s.storage.Storage[0]
		// pwd, err := os.Getwd()
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// }
		// tmp, err := template.ParseFiles(path.Join(pwd, "currentMetrics.html"))
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// }

		// err = tmp.Execute(w, s.storage.Storage)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// }
	}
}

func (s *Server) updateMetricFromBody() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err, statusCode := s.storage.UpdateMetricFromRequest(r); err != nil {
			http.Error(w, err.Error(), statusCode)
		} else {
			if s.storeInterval == 0 && s.dbPath != "" {
				if err := s.storeToFile(); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
			w.WriteHeader(statusCode)
		}
	}
}

func (s *Server) getMetricValueFromBody() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err, statusCode, value := s.storage.GetValueMetricFromBody(r); err != nil {
			http.Error(w, err.Error(), statusCode)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			w.Write(value)
			return
		}
	}
}

func (s *Server) Shutdown() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		panic(err)
	}
	// store to file
	if err := s.storeToFile(); err != nil {
		fmt.Println(err)
	}
}
