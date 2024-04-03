package pustynia

import (
	"bytes"
	"crypto/rand"
)

var DefaultCodeParams = &CodeParams{
	Alphabet:   []byte("abcdefghijklmnopqrstuvwxyz"),
	Separator:  '-',
	ChunkSize:  3,
	ChunkCount: 3,
}

type CodeParams struct {
	Alphabet   []byte
	Separator  byte
	ChunkSize  int
	ChunkCount int
}

func (p *CodeParams) size() int {
	return p.ChunkCount*p.ChunkSize + p.ChunkCount - 1
}

func (p *CodeParams) separator(i int) bool {
	return (i-p.ChunkSize)%(p.ChunkSize+1) == 0
}

func NewCode(p *CodeParams) ([]byte, error) {
	if p == nil {
		p = DefaultCodeParams
	}
	n := p.size()
	c := make([]byte, n)
	r := make([]byte, n)
	if _, err := rand.Read(r); err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		if p.separator(i) {
			c[i] = p.Separator
		} else {
			c[i] = p.Alphabet[int(r[i])%len(p.Alphabet)]
		}
	}
	return c, nil
}

func IsValidCode(p *CodeParams, c []byte) bool {
	if p == nil {
		p = DefaultCodeParams
	}
	n := p.size()
	if len(c) != n {
		return false
	}
	for i := 0; i < n; i++ {
		if p.separator(i) {
			if c[i] != p.Separator {
				return false
			}
		} else {
			if bytes.IndexByte(p.Alphabet, c[i]) == -1 {
				return false
			}
		}
	}
	return true
}
