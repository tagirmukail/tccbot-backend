package crypto

import (
	"crypto/hmac"
	"crypto/md5"  // nolint:gosec
	"crypto/sha1" // nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

const (
	HashSHA1 = iota
	HashSHA256
	HashSHA512
	HashSHA512_384
	HashMD5
)

// GetHashMessage returns a keyed-hash message authentication code using the desired
// hashtype
func GetHashMessage(hashType int, input, key []byte) []byte {
	var hasher func() hash.Hash

	switch hashType {
	case HashSHA1:
		hasher = sha1.New
	case HashSHA256:
		hasher = sha256.New
	case HashSHA512:
		hasher = sha512.New
	case HashSHA512_384:
		hasher = sha512.New384
	case HashMD5:
		hasher = md5.New
	}

	h := hmac.New(hasher, key)
	h.Write(input) // nolint:errcheck
	return h.Sum(nil)
}

// HexEncodeToString takes in a hexadecimal byte array and returns a string
func HexEncodeToString(input []byte) string {
	return hex.EncodeToString(input)
}
