package game

import (
	"crypto/sha256"
	"encoding/hex"
)

func wallHash(wall []Tile) string {
	hash := sha256.New()
	for _, tile := range wall {
		hash.Write([]byte{byte(tile)})
	}
	return hex.EncodeToString(hash.Sum(nil))
}
