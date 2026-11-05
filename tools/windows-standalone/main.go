package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Gozik standalone launcher:", err)
		os.Exit(1)
	}
}

func run() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	bundle, err := openBundle(executable)
	if err != nil {
		return err
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	id, err := bundleID(executable)
	if err != nil {
		return err
	}
	runtimeDir := filepath.Join(cacheDir, "Gozik", "standalone-"+id)
	if _, err := os.Stat(filepath.Join(runtimeDir, ".complete")); os.IsNotExist(err) {
		if err := extractAtomically(bundle, runtimeDir); err != nil {
			return err
		}
	}

	player := filepath.Join(runtimeDir, "gozik.exe")
	cmd := exec.Command(player, os.Args[1:]...)
	cmd.Dir = runtimeDir
	cmd.Env = append(os.Environ(), "GOZIK_ASSETS="+filepath.Join(runtimeDir, "assets"))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func openBundle(path string) (*zip.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	bundle, err := zip.NewReader(file, info.Size())
	if err != nil {
		file.Close()
		return nil, err
	}
	return bundle, nil
}

func bundleID(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil))[:16], nil
}

func extractAtomically(bundle *zip.Reader, destination string) error {
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}
	temporary, err := os.MkdirTemp(parent, ".extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)

	if err := extractBundle(bundle, temporary); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(temporary, ".complete"), nil, 0644); err != nil {
		return err
	}
	if err := os.Rename(temporary, destination); err != nil {
		if _, statErr := os.Stat(filepath.Join(destination, ".complete")); statErr == nil {
			return nil
		}
		return err
	}
	return nil
}

func extractBundle(bundle *zip.Reader, destination string) error {
	for _, entry := range bundle.File {
		path := filepath.Join(destination, filepath.FromSlash(entry.Name))
		relative, err := filepath.Rel(destination, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("invalid bundled path %q", entry.Name)
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
			continue
		}
		if !entry.Mode().IsRegular() {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		source, err := entry.Open()
		if err != nil {
			return err
		}
		target, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, entry.Mode())
		if err != nil {
			source.Close()
			return err
		}
		_, copyErr := io.Copy(target, source)
		closeErr := target.Close()
		source.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
