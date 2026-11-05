//go:build windows

package plugininstaller

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

func install(ctx context.Context, id string, progress Progress) error {
	archive, err := download(ctx, id, progress)
	if err != nil {
		return err
	}
	defer os.Remove(archive)
	if id == "ytmusic" {
		progress("Requesting administrator approval for the official YouTube Music installer")
		arguments := "/VERYSILENT /SUPPRESSMSGBOXES /NORESTART"
		script := "$p = Start-Process -FilePath " + psQuote(archive) + " -ArgumentList " + psQuote(arguments) + " -Verb RunAs -Wait -PassThru; exit $p.ExitCode"
		return elevatedPowerShell(ctx, progress, script)
	}
	temporary, err := os.MkdirTemp("", "gozik-spotify-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	if err := extractArchive(archive, temporary); err != nil {
		return err
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return fmt.Errorf("LOCALAPPDATA is not set")
	}
	destination := filepath.Join(localAppData, "Gozik", "Plugins", "Spotify")
	if err := os.MkdirAll(destination, 0755); err != nil {
		return err
	}
	progress("Installing Spotify, yt-dlp, and the Windows service")
	if err := copyDirectory(temporary, destination); err != nil {
		return err
	}
	script := filepath.Join(destination, "packaging", "windows", "install-service.ps1")
	progress("Requesting administrator approval for the Spotify service")
	return elevatedPowerShell(ctx, progress, "& "+psQuote(script))
}

func elevatedPowerShell(ctx context.Context, progress Progress, script string) error {
	inner := "$p = Start-Process powershell.exe -Verb RunAs -Wait -PassThru -ArgumentList '-NoProfile','-ExecutionPolicy','Bypass','-EncodedCommand','" + encodePowerShell(script) + "'; exit $p.ExitCode"
	command := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encodePowerShell(inner))
	output, err := command.CombinedOutput()
	if len(output) > 0 {
		progress(strings.TrimSpace(string(output)))
	}
	return err
}

func encodePowerShell(script string) string {
	encoded := utf16.Encode([]rune(script))
	bytes := make([]byte, len(encoded)*2)
	for i, value := range encoded {
		bytes[i*2] = byte(value)
		bytes[i*2+1] = byte(value >> 8)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func psQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func copyDirectory(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer output.Close()
		_, err = output.ReadFrom(input)
		return err
	})
}
