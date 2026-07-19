package tokenhash

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
