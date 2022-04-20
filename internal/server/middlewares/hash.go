package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/scrolllockdev/test-devops/internal/server/model"
)

func Hash(metric model.Metric, key string) (string, error) {

	h := hmac.New(sha256.New, []byte(key))

	switch metric.MType {
	case "gauge":
		fmt.Println(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), key)
		h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value)))
	case "counter":
		fmt.Println(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta), key)
		h.Write([]byte(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta)))
	default:
		return "", errors.New("bad request")
	}

	sha := hex.EncodeToString(h.Sum(nil))

	fmt.Println(sha)

	return sha, nil

}

func CheckHash(metric model.Metric, key string) bool {

	if len(metric.IncomingHash) == 0 {
		return true
	}

	hash, err := Hash(metric, key)
	if err != nil || metric.IncomingHash != hash {
		return false
	}

	return true

}
