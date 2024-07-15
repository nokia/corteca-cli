package fsutil

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func EnsureAbsolutePath(path, target string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(filepath.Join(target, path))
}

func CleanupOrCreateFolder(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.MkdirAll(path, 0755)
}

func TarAndGzip(basePath, targetTarGzPath string, includePaths []string) error {
	tarFile, err := os.Create(targetTarGzPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	gzipWriter := gzip.NewWriter(tarFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, includePath := range includePaths {
		fullPath := filepath.Join(basePath, includePath)
		if err := addPathToTarWriter(tarWriter, fullPath, basePath); err != nil {
			return err
		}

	}

	return nil
}

func addPathToTarWriter(tarWriter *tar.Writer, path, baseDir string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if file == path {
				return nil
			}

			return addFileToTarWriter(tarWriter, file, fi, path)
		})
	}
	return addFileToTarWriter(tarWriter, path, fi, baseDir)
}

func addFileToTarWriter(tarWriter *tar.Writer, file string, fi os.FileInfo, baseDir string) error {
	relPath, err := filepath.Rel(baseDir, file)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}

	header.Name = filepath.ToSlash(relPath)

	if fi.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(file)
		if err != nil {
			return err
		}
		header.Typeflag = tar.TypeSymlink
		header.Linkname = linkTarget
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	if fi.Mode().IsRegular() {
		fileContent, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fileContent.Close()

		if _, err := io.Copy(tarWriter, fileContent); err != nil {
			return err
		}
	}

	return nil
}
