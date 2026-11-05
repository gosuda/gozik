//go:build linux

package plugininstaller

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func install(ctx context.Context, id string, progress Progress) error {
	archive, err := download(ctx, id, progress)
	if err != nil {
		return err
	}
	defer os.Remove(archive)
	if id == "ytmusic" {
		progress("Installing the YouTube Music runtime and user service")
		if err := os.Chmod(archive, 0755); err != nil {
			return err
		}
		return run(ctx, progress, archive)
	}
	temporary, err := os.MkdirTemp("", "gozik-spotify-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	if err := extractArchive(archive, temporary); err != nil {
		return err
	}
	script := filepath.Join(temporary, "packaging", "linux", "install-daemon.sh")
	progress("Installing yt-dlp and registering the Spotify user service")
	if err := run(ctx, progress, "bash", script); err != nil {
		return fmt.Errorf("Spotify installation failed: %w", err)
	}
	return nil
}

func run(ctx context.Context, progress Progress, name string, args ...string) error {
	command := exec.CommandContext(ctx, name, args...)
	output, err := command.CombinedOutput()
	if len(output) > 0 {
		progress(string(output))
	}
	return err
}
