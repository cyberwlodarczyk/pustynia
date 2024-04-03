package pustynia

import "crypto/rand"

const CodeSize = 11

type Code [CodeSize]byte

func NewCode() (Code, error) {
	var c Code
	r := make([]byte, CodeSize-2)
	if _, err := rand.Read(r); err != nil {
		return c, err
	}
	for i, j := 0, 0; i < CodeSize; i++ {
		if i == 3 || i == 7 {
			c[i] = '-'
		} else {
			c[i] = 'a' + r[j]%('z'-'a'+1)
			j++
		}
	}
	return c, nil
}

func IsValidCode(c Code) bool {
	for i := 0; i < CodeSize; i++ {
		if i == 3 || i == 7 {
			if c[i] != '-' {
				return false
			}
		} else {
			if c[i] < 'a' || c[i] > 'z' {
				return false
			}
		}
	}
	return true
}
