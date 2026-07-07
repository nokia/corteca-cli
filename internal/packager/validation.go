// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package packager

import (
	"github.com/nokia/corteca-cli/internal/configuration"
	specs "github.com/nokia/corteca-cli/internal/configuration/runtimeSpec"
	"strings"
	"io"
	"bufio"
	"debug/elf"
	"fmt"
	"os"
	"path/filepath"
)

const (
	minExecutablePerm = 0111
)

var elfMachineMap = map[string]elf.Machine{
	"aarch64": elf.EM_AARCH64,
	"armv7l":  elf.EM_ARM,
	"x86_64":  elf.EM_X86_64,
}

func ValidateRootFS(rootfsPath string, targetArch string, appSettings configuration.AppSettings) error {
	
	if err := validateEntrypoint(appSettings.Entrypoint[0], targetArch, rootfsPath); err != nil {
		return err
	}

	if err := validateMounts(rootfsPath, appSettings.Runtime.Mounts); err != nil {
		return fmt.Errorf("runtime configuration error: %w", err)
	}

	return nil
}

func isBinary(entrypointPath string) (bool, string, error) {
	file, err := os.Open(entrypointPath)
	if err != nil {
		return false, "", fmt.Errorf("failed to open file: %v", err)
	}
	defer func() { _ = file.Close() }()
	is_bin := false
	const chunkSize = 4
	buf := make([]byte, chunkSize)

	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, "", err
	}

	if n < 4 {
		return false, "", fmt.Errorf("file to small")
	}

	// elf magic number 7F 45 4C 46
	if buf[0] == 0x7F && buf[1] == 0x45 && buf[2] == 0x4C && buf[3] == 0x46 {
		is_bin = true
	}

	return is_bin, file.Name(), nil
}

func isScript(entrypointPath string) (bool, string, error) {
	is_script := false
	shebang := ""
	file, err := os.Open(entrypointPath)
	if err != nil {
		return false, "", fmt.Errorf("failed to open file: %v", err)
	}
	defer func() { _ = file.Close() }()
	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, "", fmt.Errorf("failed to read first line: %v", err)
	}

	prefix := "#!"
	lineShebang := strings.TrimSpace(line)
	
	preffix_trimmed_shebang := strings.TrimPrefix(lineShebang, prefix)
	shebang = strings.SplitN(preffix_trimmed_shebang, " ", 2)[0]
	if strings.HasPrefix(lineShebang, prefix){
		is_script = true
	}

	return is_script, shebang, nil
}

func isSymlink(path string) (bool, string, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, "", fmt.Errorf("failed to resolve symlink %s: %v", path, err)
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		// It's a symlink, resolve it
		symPath, err := os.Readlink(path)
		if err != nil {
			return false, "", fmt.Errorf("failed to resolve symlink: %v", err)
		}

		return true, symPath, nil
	}
	
	return false, "", nil
}

func discoverEntrypointPath(entrypointPath, rootfs string) (string, error){
	is_symlink, sym_path, err := isSymlink(entrypointPath)
	if err != nil {
		return "", fmt.Errorf("failed to check if it's symlink %v", err)
	}

	if is_symlink {
		return discoverEntrypointPath(filepath.Join(rootfs, sym_path), rootfs) 
	}

	is_script, shebang, err := isScript(entrypointPath)
	if err != nil {
		return "", fmt.Errorf("failed to check if it's script %v", err)
	}
	
	if is_script {
		return discoverEntrypointPath(filepath.Join(rootfs, shebang), rootfs) 
	}

	isBin, bin_path, err := isBinary(entrypointPath)
	if err != nil {
		return "", fmt.Errorf("failed to check if it's binary %v", err)
	}
	
	if isBin {
		return bin_path, nil
	}

	return "", fmt.Errorf("unknown file type")
}

func validateEntrypoint(entrypoint string, targetArch, rootfsPath string) error {
	entrypointPath := filepath.Join(rootfsPath, entrypoint)
	// Validation on initial entrypoint path
	if err := validateFileProperties(entrypointPath); err != nil {
		return err
	}
	//Discover path in case of script or symlink as entrypoint
	discovered_path, err := discoverEntrypointPath(entrypointPath, rootfsPath)
	if err != nil {
		return fmt.Errorf("failed to get entrypoint path: %v", err)
	}

	if err := validateFileProperties(discovered_path); err != nil {
		return err
	}

	if err := verifyMachineCompatibility(discovered_path, targetArch); err != nil {
		return err
	}
	return nil
}

func validateFileProperties(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("entrypoint validation failed: %w", err)
	}

	if !fi.Mode().IsRegular() {
		return fmt.Errorf("entrypoint is not a regular file: %s", path)
	}

	if fi.Mode().Perm()&minExecutablePerm == 0 {
		return fmt.Errorf("entrypoint lacks execute permissions: %s", path)
	}

	return nil
}

func verifyMachineCompatibility(path, targetArch string) error {
	f, err := elf.Open(path)
	if err != nil {
		return fmt.Errorf("ELF validation failed: %w", err)
	}
	defer func() { _ = f.Close() }()

	if f.Type != elf.ET_EXEC && f.Type != elf.ET_DYN {
		return fmt.Errorf("invalid ELF type: %s", f.Type)
	}

	if f.Machine != elfMachineMap[targetArch] {
		return fmt.Errorf("architecture mismatch: binary=%s, expected=%s",
			f.Machine, elfMachineMap[targetArch])
	}

	return nil
}

func validateMounts(rootfsPath string, mounts []specs.Mount) error {
	for _, mount := range mounts {
		destination := mount.Destination

		fullPath := filepath.Join(rootfsPath, destination)
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("mount destination %q not found in rootfs", destination)
			}
			return fmt.Errorf("error checking %q: %w", destination, err)
		}
	}

	return nil
}
