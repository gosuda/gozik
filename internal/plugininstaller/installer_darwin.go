//go:build darwin

package plugininstaller

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func install(ctx context.Context, id string, progress Progress) error {
	archive, err := download(ctx, id, progress)
	if err != nil {
		return err
	}
	defer os.Remove(archive)
	if id == "ytmusic" {
		return installDMG(ctx, archive, progress)
	}
	temporary, err := os.MkdirTemp("", "gozik-spotify-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	if err := extractArchive(archive, temporary); err != nil {
		return err
	}
	destination := "/Applications/GozikSpotifyPlugin"
	progress("Requesting administrator approval to copy Spotify into /Applications")
	if err := administratorCopy(ctx, temporary, destination); err != nil {
		return err
	}
	progress("Registering the Spotify LaunchAgent; no developer toolkit is required")
	script := filepath.Join(destination, "packaging", "macos", "install-daemon.sh")
	return run(ctx, progress, "bash", script)
}

func installDMG(ctx context.Context, image string, progress Progress) error {
	mount, err := os.MkdirTemp("", "gozik-ytmusic-mount-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(mount)
	attached := false
	defer func() {
		if attached {
			_ = exec.Command("hdiutil", "detach", mount).Run()
		}
	}()
	if err := run(ctx, progress, "hdiutil", "attach", "-readonly", "-nobrowse", "-mountpoint", mount, image); err != nil {
		return err
	}
	attached = true
	progress("Requesting administrator approval to copy YouTube Music into /Applications")
	if err := administratorCopy(ctx, filepath.Join(mount, "GozikYTMusicPlugin.app"), "/Applications/GozikYTMusicPlugin.app"); err != nil {
		return err
	}
	if err := exec.Command("hdiutil", "detach", mount).Run(); err != nil {
		return err
	}
	attached = false
	progress("Registering the YouTube Music LaunchAgent; no developer toolkit is required")
	return run(ctx, progress, "bash", "/Applications/GozikYTMusicPlugin.app/Contents/MacOS/install-daemon.sh")
}

func administratorCopy(ctx context.Context, source, destination string) error {
	parent := filepath.Dir(destination)
	script := "mkdir -p " + shellQuote(parent) + " && rm -rf " + shellQuote(destination) + " && cp -R " + shellQuote(source) + " " + shellQuote(destination)
	return run(ctx, nil, "osascript", "-e", "do shell script "+appleScriptQuote(script)+" with administrator privileges")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func appleScriptQuote(value string) string {
	return "\"" + strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "\"", "\\\"") + "\""
}

func run(ctx context.Context, progress Progress, name string, args ...string) error {
	command := exec.CommandContext(ctx, name, args...)
	output, err := command.CombinedOutput()
	if progress != nil && len(output) > 0 {
		progress(strings.TrimSpace(string(output)))
	}
	if err != nil && len(output) > 0 {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return err
}
