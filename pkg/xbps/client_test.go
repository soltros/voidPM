package xbps

import (
	"testing"
)

func TestSplitPkgVersion(t *testing.T) {
	tests := []struct {
		input       string
		expectedName string
		expectedVer  string
	}{
		{"python3-gobject-3.56.2_1", "python3-gobject", "3.56.2_1"},
		{"linux-6.6.21_1", "linux", "6.6.21_1"},
		{"curl-8.6.0_1", "curl", "8.6.0_1"},
		{"simple", "simple", ""},
		{"font-sil-namdhinggo-3.100_1", "font-sil-namdhinggo", "3.100_1"},
	}

	for _, tt := range tests {
		name, ver := splitPkgVersion(tt.input)
		if name != tt.expectedName || ver != tt.expectedVer {
			t.Errorf("splitPkgVersion(%q) = (%q, %q); expected (%q, %q)",
				tt.input, name, ver, tt.expectedName, tt.expectedVer)
		}
	}
}
