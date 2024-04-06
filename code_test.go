package pustynia

import "testing"

func TestNewCode(t *testing.T) {
	for i := 0; i < 1000; i++ {
		c, err := NewCode()
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := ParseCode(c.String()); !ok {
			t.Fatalf("expected %q to be valid", c)
		}
	}
}

func TestParseCode(t *testing.T) {
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
	for _, test := range tests {
		code, got := ParseCode(test.code)
		if got != test.expected {
			if test.expected {
				t.Fatalf("expected %q to be valid", test.code)
			} else {
				t.Fatalf("expected %q to be invalid", test.code)
			}
		} else if got {
			str := code.String()
			if str != test.code {
				t.Fatalf("expected %q but got %q", test.code, str)
			}
		}
	}
}
