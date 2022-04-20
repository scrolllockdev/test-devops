package server

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	_ "github.com/lib/pq"
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
	key           string
	db            string
}

func (s *Server) Init(cfg config.Config) *Server {

	s.r = chi.NewRouter()
	s.address = cfg.ServerAddress
	s.storeInterval = cfg.StoreInterval
	s.dbPath = cfg.StorePath
	s.restore = cfg.Restore
	s.db = cfg.DBpath

	s.server = &http.Server{
		Addr:    s.address,
		Handler: s.r,
	}

	s.storage = storage.Storage{
		Storage: make([]model.Metric, 0),
	}

	s.key = cfg.Key

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
	s.r.Use(s.EqualHash)
	s.r.Post("/update/{type}/{name}/{value}", s.updateMetric())
	s.r.Get("/value/{type}/{name}", s.currentMetric())
	s.r.Get("/", s.allMetrics())
	s.r.Post("/update/", s.updateMetricFromBody())
	s.r.Post("/value/", s.getMetricValueFromBody())
	s.r.Get("/ping", s.pingDB())
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

	file, err := os.OpenFile(path.Join(pwd, s.dbPath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
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

func (s *Server) allMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Accept-Encoding", "gzip")
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		for index := range s.storage.Storage {
			item := s.storage.Storage[index]
			val := ""
			if item.MType == "gauge" {
				val = strconv.FormatFloat(*item.Value, 'e', -1, 64)
			} else {
				val = strconv.FormatInt(*item.Delta, 10)
			}

			metric := fmt.Sprintf("%s - %s - %s<br>", item.ID, item.MType, val)
			io.WriteString(w, "<html><body>"+metric+"</body></html>")
		}
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
		if err, statusCode, value := s.storage.GetValueMetricFromBody(r, s.key); err != nil {
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

func (s *Server) EqualHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var metric model.Metric

			bodyBytes, _ := ioutil.ReadAll(r.Body)
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			if len(bodyBytes) == 0 {
				next.ServeHTTP(w, r)
			}

			if err := json.Unmarshal(bodyBytes, &metric); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			hashesIsEqual := mw.CheckHash(metric, s.key)
			if !hashesIsEqual {
				http.Error(w, "status bad request", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)

	})
}

func (s *Server) pingDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("postgres",
			"db.db")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = db.PingContext(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
