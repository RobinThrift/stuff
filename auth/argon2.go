package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type Argon2Params struct {
	KeyLen  uint32 `json:"keyLen"`
	Memory  uint32 `json:"memory"`
	Threads uint8  `json:"threads"`
	Time    uint32 `json:"time"`
	Version int    `json:"version"`
}

func (a *Argon2Params) toJSONString() (string, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return "", fmt.Errorf("error marshalling argon2 prams to JSON: %w", err)
	}
	return string(b), nil
}

func checkPassword(candidatePlain []byte, hash []byte, salt, paramsJSON []byte) (bool, error) {
	var params Argon2Params

	err := json.Unmarshal(paramsJSON, &params)
	if err != nil {
		return false, err
	}

	candidate := argon2.IDKey(candidatePlain, salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	return subtle.ConstantTimeCompare(hash, candidate) == 1, nil
}

func encryptPassword(plaintextPasswd []byte, params Argon2Params) ([]byte, []byte, error) {
	var salt [16]byte
	_, err := rand.Read(salt[:])
	if err != nil {
		return nil, nil, fmt.Errorf("error generating random salt: %w", err)
	}

	return argon2.IDKey(plaintextPasswd, salt[:], params.Time, params.Memory, params.Threads, params.KeyLen), salt[:], nil
}
