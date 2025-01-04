package password

import (
	"runtime"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

var (
	DefaultPolicy = &Policy{
		Upper:     1,
		Lower:     1,
		Number:    1,
		Special:   1,
		MinLength: 12,
		MaxLength: 64,
	}
	DefaultParams = &Params{
		Memory:     64 * 1024,
		Iterations: 1,
		Threads:    uint8(runtime.NumCPU()),
		KeyLength:  32,
	}
)

type Policy struct {
	Upper     int
	Lower     int
	Number    int
	Special   int
	MinLength int
	MaxLength int
}

type Params struct {
	Memory     uint32
	Iterations uint32
	Threads    uint8
	KeyLength  uint32
}

func IsValid(password []byte, policy *Policy) bool {
	var (
		r       rune
		size    int
		upper   int
		lower   int
		number  int
		special int
	)
	for i := 0; i < len(password); {
		r, size = utf8.DecodeRune(password[i:])
		switch {
		case unicode.IsUpper(r):
			upper++
		case unicode.IsLower(r):
			lower++
		case unicode.IsNumber(r):
			number++
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			special++
		}
		i += size
	}
	return upper >= policy.Upper &&
		lower >= policy.Lower &&
		number >= policy.Number &&
		special >= policy.Special &&
		len(password) >= policy.MinLength &&
		len(password) <= policy.MaxLength
}

func Hash(password []byte, salt []byte, params *Params) []byte {
	return argon2.IDKey(
		password,
		salt,
		params.Iterations,
		params.Memory,
		params.Threads,
		params.KeyLength,
	)
}
