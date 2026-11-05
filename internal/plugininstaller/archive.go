package plugininstaller

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func extractArchive(archivePath, destination string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, destination)
	}
	return extractTarGz(archivePath, destination)
}

func extractZip(archivePath, destination string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	for _, entry := range archive.File {
		path, err := archivePathFor(destination, entry.Name)
		if err != nil {
			return err
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
		source, err := entry.Open()
		if err != nil {
			return err
		}
		if err := writeArchiveFile(path, source, entry.Mode()); err != nil {
			source.Close()
			return err
		}
		source.Close()
	}
	return nil
}

func extractTarGz(archivePath, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer compressed.Close()
	archive := tar.NewReader(compressed)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		path, err := archivePathFor(destination, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := writeArchiveFile(path, archive, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}
}

func archivePathFor(destination, name string) (string, error) {
	path := filepath.Join(destination, filepath.FromSlash(name))
	relative, err := filepath.Rel(destination, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid archive path %q", name)
	}
	return path, nil
}

func writeArchiveFile(path string, source io.Reader, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if mode.Perm() == 0 {
		mode = 0644
	}
	target, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(target, source)
	closeErr := target.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
