package code

import (
	"crypto/rand"
	"fmt"
)

const Size = 11

type Code [Size]byte

func (c Code) String() string {
	return string(c.Bytes())
}

func (c Code) Bytes() []byte {
	return c[:]
}

func New() (Code, error) {
	var c Code
	r := make([]byte, Size-2)
	if _, err := rand.Read(r); err != nil {
		return c, fmt.Errorf("error generating random numbers: %w", err)
	}
	for i, j := 0, 0; i < Size; i++ {
		if i == 3 || i == 7 {
			c[i] = '-'
		} else {
			c[i] = 'a' + r[j]%('z'-'a'+1)
			j++
		}
	}
	return c, nil
}

func Parse(s string) (Code, bool) {
	var c Code
	if len(s) != Size {
		return c, false
	}
	for i := 0; i < Size; i++ {
		if i == 3 || i == 7 {
			if s[i] != '-' {
				return c, false
			}
		} else {
			if s[i] < 'a' || s[i] > 'z' {
				return c, false
			}
		}
	}
	copy(c[:], s)
	return c, true
}
