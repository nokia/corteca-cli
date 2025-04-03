// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

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

	destDir := filepath.Dir(dst)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return 0, err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	if err != nil {
		return nBytes, err
	}

	if err := os.Chmod(dst, sourceFileStat.Mode()); err != nil {
		return nBytes, err
	}
	return nBytes, nil
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

func EnsureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
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

func ExtractTarball(src, dest string) error {
	// Open the gzip file
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open source file: %v", err)
	}
	defer file.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("could not create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Iterate through the tar entries
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return fmt.Errorf("could not read tar entry: %v", err)
		}

		// Determine the proper path for the file
		targetPath := filepath.Join(dest, header.Name)

		// Handle directories and regular files
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("could not create directory: %v", err)
			}
		case tar.TypeReg:
			dir := filepath.Dir(targetPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("could not create directory: %v", err)
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("could not create file: %v", err)
			}
			defer outFile.Close()

			err = os.Chmod(targetPath, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("could not set permissions to file: %v", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("could not write file: %v", err)
			}
		case tar.TypeSymlink:
			linkTarget := header.Linkname

			if err := os.Symlink(linkTarget, targetPath); err != nil {
				return fmt.Errorf("failed to create symlink: %w", err)
			}
		case tar.TypeLink:
			linkTarget := filepath.Join(dest, header.Linkname)

			if err := os.Link(linkTarget, targetPath); err != nil {
				return fmt.Errorf("failed to create hard link: %w", err)
			}
		default:
			return fmt.Errorf("unknown tar entry type: %v", header.Typeflag)
		}
	}

	return nil
}

func CreateTarArchive(fileName, archivePath string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	archive, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer archive.Close()

	writer := tar.NewWriter(archive)
	defer writer.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	header := &tar.Header{
		Name: filepath.Base(fileName),
		Mode: 0644,
		Size: fileInfo.Size(),
	}

	if err := writer.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err = io.Copy(writer, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

func RemoveFilesFromFolder(directoryPath string, filesToRemove []string) error {
	for _, fileName := range filesToRemove {
		filePath := filepath.Join(directoryPath, fileName)
		if err := os.RemoveAll(filePath); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove %s: %v", filePath, err)
			}
		}
	}
	return nil
}

func CreateTarArchiveFromFolder(srcDir, destTarGzPath string) error {
	archive, err := os.Create(destTarGzPath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer archive.Close()

	gzipWriter := gzip.NewWriter(archive)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	baseDir := filepath.Clean(srcDir)
	return filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if file == srcDir {
			return nil // Skip the top-level directory
		}

		relPath, err := filepath.Rel(baseDir, file)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil // Skip directories
		}

		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

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
	})
}
