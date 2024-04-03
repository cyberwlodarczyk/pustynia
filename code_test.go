package pustynia

import "testing"

func TestNewCode(t *testing.T) {
	for i := 0; i < 1000; i++ {
		c, err := NewCode()
		if err != nil {
			t.Fatal(err)
		}
		if !IsValidCode(c) {
			t.Fatalf("expected %q to be valid", c)
		}
	}
}

func TestIsValidCode(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{"abc-def-ghi", true},
		{"zyx-wvu-tsr", true},
		{"oec-mao/oap", false},
		{"orj#tow-soc", false},
		{"sdo-wpdapmw", false},
		{"pqopscodnso", false},
		{"-spdj-sodsm", false},
	}
	var c Code
	for _, test := range tests {
		copy(c[:], test.code)
		got := IsValidCode(c)
		if got != test.expected {
			if test.expected {
				t.Fatalf("expected %q to be valid", test.code)
			} else {
				t.Fatalf("expected %q to be invalid", test.code)
			}
		}
	}
}
