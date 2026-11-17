//go:build linux || freebsd || openbsd || netbsd

package platforms

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallIconsPrefersPayloadIconTreeOverPreview(t *testing.T) {
	home := t.TempDir()
	installDir := t.TempDir()
	t.Setenv("HOME", home)
	iconDir := filepath.Join(installDir, "assets", "icons", "hicolor", "256x256", "apps")
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		t.Fatal(err)
	}
	want := []byte("icon payload")
	if err := os.WriteFile(filepath.Join(iconDir, "gozik.png"), want, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installDir, "gozik.png"), []byte("preview payload"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := installIcons(&InstallContext{InstallDir: installDir}, false); err != nil {
		t.Fatalf("installIcons: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(home, ".local", "share", "icons", "hicolor", "256x256", "apps", "gozik.png"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("installed icon = %q, want %q", got, want)
	}
}

func TestInstallIconsFallsBackWhenPayloadIconTreeIsEmpty(t *testing.T) {
	home := t.TempDir()
	installDir := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(installDir, "assets", "icons", "hicolor", "256x256", "apps"), 0755); err != nil {
		t.Fatal(err)
	}
	want := []byte("png payload")
	if err := os.WriteFile(filepath.Join(installDir, "gozik.png"), want, 0644); err != nil {
		t.Fatal(err)
	}

	if err := installIcons(&InstallContext{InstallDir: installDir}, true); err != nil {
		t.Fatalf("installIcons: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(home, ".local", "share", "icons", "hicolor", "256x256", "apps", "gozik.png"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("installed icon = %q, want %q", got, want)
	}
}

func TestExtractAppImageIconCreatesWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	appImage := filepath.Join(root, "gozik.AppImage")
	script := "#!/bin/sh\nmkdir -p squashfs-root/usr/share/icons/hicolor/256x256/apps\nprintf icon > squashfs-root/usr/share/icons/hicolor/256x256/apps/gozik.png\n"
	if err := os.WriteFile(appImage, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	extractDir := filepath.Join(root, "missing", "extract")
	iconsDst := filepath.Join(root, "icons")

	if err := extractAppImageIcon(appImage, extractDir, iconsDst); err != nil {
		t.Fatalf("extractAppImageIcon: %v", err)
	}
	if _, err := os.Stat(filepath.Join(iconsDst, "256x256", "apps", "gozik.png")); err != nil {
		t.Fatal(err)
	}
}
