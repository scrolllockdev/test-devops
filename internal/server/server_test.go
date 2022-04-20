package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/scrolllockdev/test-devops/internal/server/storage"
)

func TestServerRestoreFromFile(t *testing.T) {
	type fields struct {
		r             *chi.Mux
		server        *http.Server
		address       string
		storeInterval time.Duration
		dbPath        string
		restore       bool
		storage       storage.Storage
	}
	type args struct {
		storage *storage.Storage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				r:             tt.fields.r,
				server:        tt.fields.server,
				address:       tt.fields.address,
				storeInterval: tt.fields.storeInterval,
				dbPath:        tt.fields.dbPath,
				restore:       tt.fields.restore,
				storage:       tt.fields.storage,
			}
			if err := s.restoreFromFile(tt.args.storage); (err != nil) != tt.wantErr {
				t.Errorf("Server.restoreFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
