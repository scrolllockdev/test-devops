package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/scrolllockdev/test-devops/internal/model"
)

func EqualHashes(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
				hashesIsEqual := CheckHash(metric, key)
				if !hashesIsEqual {
					http.Error(w, "status bad request", http.StatusBadRequest)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Hash(metric model.Metric, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	switch metric.MType {
	case "gauge":
		h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value)))
	case "counter":
		h.Write([]byte(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta)))
	default:
		return "", errors.New("bad request")
	}
	sha := hex.EncodeToString(h.Sum(nil))
	return sha, nil

}

func CheckHash(metric model.Metric, key string) bool {
	switch len(metric.Hash) {
	case 0:
		if len(key) == 0 {
			return true
		} else {
			return false
		}
	default:
		if len(key) == 0 {
			return true
		} else {
			hash, err := Hash(metric, key)
			if err != nil || metric.Hash != hash {
				return false
			}
			return true
		}
	}
}
