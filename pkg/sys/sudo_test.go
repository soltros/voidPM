package sys

import (
	"testing"
)

func TestIsRoot(t *testing.T) {
	_ = IsRoot()
}

func TestWrapElevated(t *testing.T) {
	_, err := WrapElevated([]string{})
	if err == nil {
		t.Errorf("expected error for empty args slice, got nil")
	}

	args, err := WrapElevated([]string{"ls", "-la"})
	if err != nil && IsRoot() {
		t.Errorf("unexpected error for root user: %v", err)
	}
	if len(args) == 0 {
		t.Errorf("expected non-empty wrapped args slice")
	}
}

func TestEnsurePrivileges(t *testing.T) {
	err := EnsurePrivileges()
	if IsRoot() && err != nil {
		t.Errorf("unexpected error for root user in EnsurePrivileges: %v", err)
	}
	if !IsRoot() && err == nil {
		t.Errorf("expected error for non-root user in EnsurePrivileges")
	}
}
