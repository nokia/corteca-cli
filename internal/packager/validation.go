// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package packager

import (
	"corteca/internal/configuration"
	specs "corteca/internal/configuration/runtimeSpec"
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
	entrypointPath := filepath.Join(rootfsPath, appSettings.Entrypoint)
	if err := validateEntrypoint(entrypointPath, targetArch); err != nil {
		return err
	}

	if err := validateMounts(rootfsPath, appSettings.Runtime.Mounts); err != nil {
		return fmt.Errorf("runtime configuration error: %w", err)
	}

	return nil
}

func validateEntrypoint(entrypointPath string, targetArch string) error {
	if err := validateFileProperties(entrypointPath); err != nil {
		return err
	}

	if err := verifyMachineCompatibility(entrypointPath, targetArch); err != nil {
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
	defer f.Close()

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
