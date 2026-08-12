package update

import (
	"testing"
)

func TestNewUpdaterDefaults(t *testing.T) {
	u := NewUpdater("", "", "")
	if u.Repo != DefaultGitHubRepo {
		t.Errorf("Expected Repo %q, got %q", DefaultGitHubRepo, u.Repo)
	}
	if u.Branch != DefaultGitHubBranch {
		t.Errorf("Expected Branch %q, got %q", DefaultGitHubBranch, u.Branch)
	}
	if u.TargetPath != DefaultTargetPath {
		t.Errorf("Expected TargetPath %q, got %q", DefaultTargetPath, u.TargetPath)
	}
}

func TestGetRawBinaryURL(t *testing.T) {
	u := NewUpdater("soltros/voidPM", "main", "/usr/bin/vpm")

	url := u.GetRawBinaryURL()
	expected := "https://raw.githubusercontent.com/soltros/voidPM/main/vpm"
	if url != expected {
		t.Errorf("Expected URL %q, got %q", expected, url)
	}
}
