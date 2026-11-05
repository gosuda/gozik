package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAppendedBundle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gozik.exe")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("launcher")); err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	entry, err := archive.Create("gozik.exe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("player")); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	bundle, err := openBundle(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.File) != 1 || bundle.File[0].Name != "gozik.exe" {
		t.Fatalf("unexpected bundle: %+v", bundle.File)
	}
	reader, err := bundle.File[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "player" {
		t.Fatalf("bundled player = %q", content)
	}
}

func TestExtractBundleRejectsTraversal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	if _, err := archive.Create("../outside"); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	bundle, err := openBundle(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := extractBundle(bundle, t.TempDir()); err == nil {
		t.Fatal("extractBundle accepted a traversal path")
	}
}
