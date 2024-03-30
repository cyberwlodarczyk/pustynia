package argon2id

import (
	"crypto/rand"
	"runtime"

	"golang.org/x/crypto/argon2"
)

const Version = argon2.Version

var DefaultParams = &Params{
	Memory:      64 * 1024,
	Iterations:  1,
	Parallelism: uint8(runtime.NumCPU()),
	SaltLength:  16,
	KeyLength:   32,
}

type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func Key(p *Params, salt, password []byte) []byte {
	return argon2.IDKey(
		password,
		salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		p.KeyLength,
	)
}

func RandomSalt(p *Params) ([]byte, error) {
	b := make([]byte, p.SaltLength)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
