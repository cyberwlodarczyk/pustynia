package pustynia

import "testing"

var (
	cp1, cp2, cp3 = &CodeParams{
		Alphabet:   []byte("0123456789"),
		Separator:  '/',
		ChunkSize:  2,
		ChunkCount: 4,
	}, &CodeParams{
		Alphabet:   []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
		Separator:  '_',
		ChunkSize:  4,
		ChunkCount: 3,
	}, &CodeParams{
		Alphabet:   []byte(";:,.'\""),
		Separator:  '$',
		ChunkSize:  1,
		ChunkCount: 3,
	}
)

func TestNewCode(t *testing.T) {
	tests := []struct {
		params *CodeParams
	}{
		{nil},
		{cp1},
		{cp2},
		{cp3},
	}
	for _, test := range tests {
		for i := 0; i < 10; i++ {
			c, err := NewCode(test.params)
			if err != nil {
				t.Fatal(err)
			}
			if !IsValidCode(test.params, c) {
				t.Fatalf("expected %q to be valid", c)
			}
		}
	}
}

func TestIsValidCode(t *testing.T) {
	tests := []struct {
		params   *CodeParams
		code     string
		expected bool
	}{
		{nil, "abc-def-ghi", true},
		{nil, "zyx-wvu/tsr", false},
		{nil, "abcd-de-ghi", false},
		{nil, "abd-def-ghi-jkl", false},
		{cp1, "98/77/43/11", true},
		{cp1, "983-012-302", false},
		{cp2, "ZDOE_KFOW_QPAJ", true},
		{cp2, "PRKW-CNIE_NSOF", false},
		{cp3, ";$:$'", true},
		{cp3, "$$$''", false},
	}
	for _, test := range tests {
		got := IsValidCode(test.params, []byte(test.code))
		if got != test.expected {
			if test.expected {
				t.Fatalf("expected %q to be valid", test.code)
			} else {
				t.Fatalf("expected %q to be invalid", test.code)
			}
		}
	}
}
