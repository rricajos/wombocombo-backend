package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
)

const roomCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no ambiguous chars

// GenerateRoomCode generates a 6-char alphanumeric room code.
func GenerateRoomCode() string {
	var sb strings.Builder
	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(roomCodeChars))))
		sb.WriteByte(roomCodeChars[n.Int64()])
	}
	return sb.String()
}

// GenerateRandomHex returns a random hex string of the given byte length.
func GenerateRandomHex(byteLen int) string {
	b := make([]byte, byteLen)
	rand.Read(b)
	return hex.EncodeToString(b)
}
