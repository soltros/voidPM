package runit

import (
	"testing"
)

func TestFormattedUptime(t *testing.T) {
	srv := &Service{UptimeSeconds: 45}
	if srv.FormattedUptime() != "45s" {
		t.Errorf("expected '45s', got %q", srv.FormattedUptime())
	}

	srv.UptimeSeconds = 125
	if srv.FormattedUptime() != "2m 5s" {
		t.Errorf("expected '2m 5s', got %q", srv.FormattedUptime())
	}

	srv.UptimeSeconds = 7320
	if srv.FormattedUptime() != "2h 2m" {
		t.Errorf("expected '2h 2m', got %q", srv.FormattedUptime())
	}

	srv.UptimeSeconds = 90000
	if srv.FormattedUptime() != "1d 1h" {
		t.Errorf("expected '1d 1h', got %q", srv.FormattedUptime())
	}
}

func TestFormatLogLineTAI64N(t *testing.T) {
	input := "@4000000065c92c8012345678 Service started successfully"
	formatted := FormatLogLine(input)
	if formatted == input {
		t.Errorf("expected TAI64N timestamp to be formatted, got original %q", formatted)
	}

	plain := "Just a normal log line"
	if FormatLogLine(plain) != plain {
		t.Errorf("expected plain log line to remain unchanged, got %q", FormatLogLine(plain))
	}
}
