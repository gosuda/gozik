//go:build !linux && !windows && !darwin

package plugininstaller

import (
	"context"
	"fmt"
	"runtime"
)

func install(context.Context, string, Progress) error {
	return fmt.Errorf("automatic plugin installation is not supported on %s", runtime.GOOS)
}
