package domain

import (
	"crypto/rand"
	"errors"
)

// roomCodeAlphabet is unambiguous: no 0/O, 1/I/L. Uppercase A-Z minus I,L,O
// and digits 2-9.
const roomCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// GenerateRoomCode returns a short uppercase join code (default 4 chars) from
// an unambiguous alphabet, e.g. "AB12".
func GenerateRoomCode() (string, error) {
	return GenerateRoomCodeN(4)
}

// GenerateRoomCodeN returns a code of exactly n characters.
func GenerateRoomCodeN(n int) (string, error) {
	if n <= 0 || n > 16 {
		return "", errors.New("invalid room code length")
	}
	out := make([]byte, n)
	if _, err := rand.Read(out); err != nil {
		return "", err
	}
	for i, b := range out {
		out[i] = roomCodeAlphabet[int(b)%len(roomCodeAlphabet)]
	}
	return string(out), nil
}
