package pustynia

import "crypto/rand"

const CodeSize = 11

type Code [CodeSize]byte

func (c Code) String() string {
	return string(c.Bytes())
}

func (c Code) Bytes() []byte {
	return c[:]
}

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

func ParseCode(s string) (Code, bool) {
	var c Code
	if len(s) != CodeSize {
		return c, false
	}
	for i := 0; i < CodeSize; i++ {
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
