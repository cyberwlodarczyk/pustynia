package id

import (
	"math/rand"
	"strings"
)

var DefaultParams = &Params{
	Alphabet:   "abcdefghijklmnopqrstuvwxyz",
	Separator:  '-',
	ChunkSize:  3,
	ChunkCount: 3,
}

type Params struct {
	Alphabet   string
	Separator  byte
	ChunkSize  int
	ChunkCount int
}

func (p *Params) size() int {
	return p.ChunkCount*p.ChunkSize + p.ChunkCount - 1
}

func (p *Params) separator(i int) bool {
	return (i-p.ChunkSize)%(p.ChunkSize+1) == 0
}

func New(p *Params) string {
	var sb strings.Builder
	for i := 0; i < p.size(); i++ {
		if p.separator(i) {
			sb.WriteByte(p.Separator)
		} else {
			sb.WriteByte(p.Alphabet[rand.Intn(len(p.Alphabet))])
		}
	}
	return sb.String()
}

func IsValid(p *Params, s string) bool {
	n := p.size()
	if len(s) != n {
		return false
	}
	for i := 0; i < n; i++ {
		if p.separator(i) {
			if s[i] != p.Separator {
				return false
			}
		} else {
			if strings.IndexByte(p.Alphabet, s[i]) == -1 {
				return false
			}
		}
	}
	return true
}
