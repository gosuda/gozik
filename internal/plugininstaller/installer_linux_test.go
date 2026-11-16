//go:build linux

package plugininstaller

import (
	"slices"
	"strings"
	"testing"
)

func TestHostEnvironmentExcludesAppImageLinkerOverrides(t *testing.T) {
	t.Setenv("LD_LIBRARY_PATH", "/tmp/gozik-appimage/usr/lib")
	t.Setenv("LD_PRELOAD", "/tmp/gozik-appimage/usr/lib/libpreload.so")
	t.Setenv("GOZIK_TEST_ENV", "preserved")

	environment := hostEnvironment()
	for _, variable := range environment {
		name, _, _ := strings.Cut(variable, "=")
		if name == "LD_LIBRARY_PATH" || name == "LD_PRELOAD" {
			t.Fatalf("hostEnvironment retained %s", variable)
		}
	}
	if !slices.Contains(environment, "GOZIK_TEST_ENV=preserved") {
		t.Fatal("hostEnvironment removed an unrelated variable")
	}
}
