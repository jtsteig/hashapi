package main

import (
	"crypto/sha512"
	"encoding/base64"
	"time"
)

// CalculateHash returns the Sha512 and base64 encoding of a string.
func CalculateHash(hashCount string) (string, time.Duration) {
	start := time.Now()
	sha512 := sha512.New()
	sha512.Write([]byte(hashCount))
	ret := base64.StdEncoding.EncodeToString(sha512.Sum(nil))
	return ret, time.Since(start)
}
