package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Hash(msg string, key string) string {

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	sha := hex.EncodeToString(h.Sum(nil))

	return sha
}
