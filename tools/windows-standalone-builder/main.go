package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	runtimeDir := flag.String("runtime", "", "directory containing the Windows runtime")
	launcher := flag.String("launcher", "", "standalone launcher executable")
	output := flag.String("output", "", "output standalone executable")
	flag.Parse()
	if *runtimeDir == "" || *launcher == "" || *output == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := pack(*runtimeDir, *launcher, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func pack(runtimeDir, launcher, output string) error {
	source, err := os.Open(launcher)
	if err != nil {
		return err
	}
	defer source.Close()
	target, err := os.Create(output)
	if err != nil {
		return err
	}
	if _, err := io.Copy(target, source); err != nil {
		target.Close()
		return err
	}

	archive := zip.NewWriter(target)
	err = filepath.WalkDir(runtimeDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(runtimeDir, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relative)
		header.Method = zip.Deflate
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if err != nil {
		archive.Close()
		target.Close()
		return err
	}
	if err := archive.Close(); err != nil {
		target.Close()
		return err
	}
	return target.Close()
}
