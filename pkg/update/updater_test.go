package update

import (
	"runtime"
	"testing"
)

func TestNewUpdaterDefaults(t *testing.T) {
	u := NewUpdater("", "")
	if u.Repo != DefaultGitHubRepo {
		t.Errorf("Expected Repo %q, got %q", DefaultGitHubRepo, u.Repo)
	}
	if u.TargetPath != DefaultTargetPath {
		t.Errorf("Expected TargetPath %q, got %q", DefaultTargetPath, u.TargetPath)
	}
}

func TestFindBinaryURL(t *testing.T) {
	u := NewUpdater("soltros/voidPM", "/usr/bin/vpm")

	rel := &ReleaseInfo{
		TagName: "v0.2.0",
		Assets: []ReleaseAsset{
			{Name: "README.md", BrowserDownloadURL: "https://example.com/readme"},
			{Name: "vpm-linux-" + runtime.GOARCH, BrowserDownloadURL: "https://example.com/vpm-bin"},
		},
	}

	url := u.FindBinaryURL(rel)
	expected := "https://example.com/vpm-bin"
	if url != expected {
		t.Errorf("Expected URL %q, got %q", expected, url)
	}
}
