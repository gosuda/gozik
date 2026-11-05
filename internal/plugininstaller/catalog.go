package plugininstaller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const PluginBundleVersion = "plugin-v0.6"

type Plugin struct {
	ID          string
	Name        string
	Version     string
	Description string
	Repository  string
	WebUI       string
}

var Plugins = []Plugin{
	{ID: "spotify", Name: "Spotify", Version: "v0.6", Description: "Spotify search and playback with yt-dlp included.", Repository: "https://github.com/gosuda/gozik-spotify", WebUI: "http://127.0.0.1:50055"},
	{ID: "ytmusic", Name: "YouTube Music", Version: "v0.5", Description: "YouTube Music with its Python and Node.js runtime included.", Repository: "https://github.com/gosuda/gozik-yt-music", WebUI: "http://127.0.0.1:50053"},
}

type Progress func(message string)

type release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type asset struct {
	name string
	url  string
}

func Lookup(id string) (Plugin, bool) {
	for _, plugin := range Plugins {
		if plugin.ID == id {
			return plugin, true
		}
	}
	return Plugin{}, false
}

func Install(ctx context.Context, id string, progress Progress) error {
	if _, ok := Lookup(id); !ok {
		return fmt.Errorf("unknown plugin %q", id)
	}
	return install(ctx, id, progress)
}

func download(ctx context.Context, id string, progress Progress) (string, error) {
	selected, err := selectAsset(ctx, id, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	progress("Downloading " + selected.name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, selected.url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "gozik-plugin-installer/1.0")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", response.Status)
	}
	file, err := os.CreateTemp("", "gozik-plugin-*"+filepath.Ext(selected.name))
	if err != nil {
		return "", err
	}
	name := file.Name()
	if _, err := io.Copy(&progressWriter{writer: file, progress: progress}, response.Body); err != nil {
		file.Close()
		os.Remove(name)
		return "", err
	}
	if err := file.Close(); err != nil {
		os.Remove(name)
		return "", err
	}
	return name, nil
}

func selectAsset(ctx context.Context, id, goos, goarch string) (asset, error) {
	plugin, ok := Lookup(id)
	if !ok {
		return asset{}, fmt.Errorf("unknown plugin %q", id)
	}
	url := "https://api.github.com/repos/gosuda/gozik/releases/tags/" + PluginBundleVersion
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return asset{}, err
	}
	req.Header.Set("User-Agent", "gozik-plugin-installer/1.0")
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return asset{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return asset{}, fmt.Errorf("GitHub release lookup failed for %s: %s", PluginBundleVersion, response.Status)
	}
	var release release
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return asset{}, err
	}
	prefix, extension := assetPattern(id, goos, goarch)
	for _, candidate := range release.Assets {
		if strings.HasPrefix(candidate.Name, prefix) && strings.HasSuffix(candidate.Name, extension) {
			return asset{name: candidate.Name, url: candidate.BrowserDownloadURL}, nil
		}
	}
	return asset{}, fmt.Errorf("no %s %s package for %s/%s in %s", plugin.Name, plugin.Version, goos, goarch, PluginBundleVersion)
}

func assetPattern(id, goos, goarch string) (string, string) {
	platform := goos + "-" + goarch
	if goos == "darwin" {
		platform = "macos-" + goarch
	}
	if id == "ytmusic" {
		switch goos {
		case "windows":
			return "gozik-ytmusic-plugin-" + platform, ".exe"
		case "darwin":
			return "gozik-ytmusic-plugin-" + platform + "-", ".dmg"
		default:
			return "gozik-ytmusic-plugin-" + platform + "-", ".AppImage"
		}
	}
	extension := ".tar.gz"
	if goos == "windows" {
		extension = ".zip"
	}
	return "gozik-spotify-plugin-" + platform + "-", extension
}

type progressWriter struct {
	writer   io.Writer
	progress Progress
	written  int64
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.written += int64(n)
	if w.written/(10*1024*1024) != (w.written-int64(n))/(10*1024*1024) {
		w.progress(fmt.Sprintf("Downloaded %d MiB", w.written/(1024*1024)))
	}
	return n, err
}
