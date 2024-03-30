package id

import "testing"

var (
	p1, p2, p3 = &Params{
		Alphabet:   "0123456789",
		Separator:  '/',
		ChunkSize:  2,
		ChunkCount: 4,
	}, &Params{
		Alphabet:   "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		Separator:  '_',
		ChunkSize:  4,
		ChunkCount: 3,
	}, &Params{
		Alphabet:   ";:,.'\"",
		Separator:  '$',
		ChunkSize:  1,
		ChunkCount: 3,
	}
)

func TestNew(t *testing.T) {
	tests := []struct {
		params *Params
	}{
		{DefaultParams},
		{p1},
		{p2},
		{p3},
	}
	for _, test := range tests {
		for i := 0; i < 10; i++ {
			id := New(test.params)
			if !IsValid(test.params, id) {
				t.Fatalf("expected %q to be valid", id)
			}
		}
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		params   *Params
		id       string
		expected bool
	}{
		{DefaultParams, "abc-def-ghi", true},
		{DefaultParams, "zyx-wvu/tsr", false},
		{DefaultParams, "abcd-de-ghi", false},
		{DefaultParams, "abd-def-ghi-jkl", false},
		{p1, "98/77/43/11", true},
		{p1, "983-012-302", false},
		{p2, "ZDOE_KFOW_QPAJ", true},
		{p2, "PRKW-CNIE_NSOF", false},
		{p3, ";$:$'", true},
		{p3, "$$$''", false},
	}
	for _, test := range tests {
		got := IsValid(test.params, test.id)
		if got != test.expected {
			if test.expected {
				t.Fatalf("expected %q to be valid", test.id)
			} else {
				t.Fatalf("expected %q to be invalid", test.id)
			}
		}
	}
}
