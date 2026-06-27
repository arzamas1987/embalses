package version

import "testing"

func TestGet(t *testing.T) {
	info := Get()
	if info.Version == "" {
		t.Error("expected non-empty version")
	}
	if info.Go == "unknown" {
		t.Skip("build info not available in test mode")
	}
}
