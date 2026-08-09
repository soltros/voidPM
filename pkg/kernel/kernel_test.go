package kernel

import (
	"testing"
)

func TestKernelPkgRegex(t *testing.T) {
	valid := []string{"linux", "linux6.6", "linux-lts", "linux-mainline", "linux6.1"}
	invalid := []string{"linux-firmware", "nvidia-dkms", "curl", "linux-headers"}

	for _, v := range valid {
		if !kernelPkgRegex.MatchString(v) {
			t.Errorf("expected %q to match kernel package regex", v)
		}
	}

	for _, inv := range invalid {
		if kernelPkgRegex.MatchString(inv) {
			t.Errorf("expected %q NOT to match kernel package regex", inv)
		}
	}
}
