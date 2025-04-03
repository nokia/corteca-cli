// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package packager

import (
	"corteca/internal/configuration"
	"corteca/internal/fsutil"
	"fmt"
	"os"
	"path/filepath"
)

const (
	runtimeConfigFile            = "config.json"
	runtimeConfigTarFile		 = runtimeConfigFile + ".tar"
	adfFile                      = "ADF"
	nokiaRuntimeConfigAnnotation = "com.nokia.runtime.config"
	buildInfoPath                = "buildinfo.json"
)

func AnnotateRootFS(dest string, appSettings configuration.AppSettings, buildMetadata map[string]string) error {
	if len(buildMetadata) > 0 && appSettings.Runtime.Annotations == nil {
		appSettings.Runtime.Annotations = make(map[string]string)
	}

	for key, value := range buildMetadata {
		appSettings.Runtime.Annotations[key] = value
	}

	if err := writeBuildInfoJson(buildMetadata, dest); err != nil {
		return fmt.Errorf("failed to write build info: %v", err)
	}

	return nil
}

func PackageOCI(buildDir, distPath, arch, platform, rootfsTarGzPath string, appSettings configuration.AppSettings) error {
	ociDirName := fmt.Sprintf("%s-%s-%s-oci", appSettings.Name, appSettings.Version, arch)
	ociTarName := fmt.Sprintf("%s-%s-%s-oci.tar", appSettings.Name, appSettings.Version, arch)
	ociDirPath := filepath.Join(buildDir, ociDirName)
	ociTarPath := filepath.Join(distPath, ociTarName)

	// break to build and package
	if err := createOCILayout(ociDirPath, rootfsTarGzPath, filepath.Join(buildDir, runtimeConfigTarFile), platform, appSettings); err != nil {
		return fmt.Errorf("failed to create OCI layout: %v", err)
	}

	if err := fsutil.CreateTarArchiveFromFolder(ociDirPath, ociTarPath); err != nil {
		return fmt.Errorf("failed to create OCI tar: %v", err)
	}

	return nil
}

// refers to legacy rootfs
func PackageRootFS(buildDir, appDir, distPath, arch, outputType string, appSettings configuration.AppSettings) error {

	rootfsTarGz := "rootfs.tar.gz" // Name for the rootfs archive
	appPackage := fmt.Sprintf("%s-%s-%s-%s.tar.gz", appSettings.Name, appSettings.Version, arch, outputType)

	_, err := fsutil.CopyFile(filepath.Join(appDir, adfFile), filepath.Join(buildDir, adfFile))
	if err != nil {
		return fmt.Errorf("failed to copy ADF file - build dir: %v", err)
	}
	err = os.Chmod(filepath.Join(buildDir, adfFile), 0666)
	if err != nil {
		return fmt.Errorf("failed to change permissions to temp ADF '%s': %v", filepath.Join(buildDir, adfFile), err)
	}
	if err := fsutil.TarAndGzip(buildDir, filepath.Join(distPath, appPackage), []string{adfFile, rootfsTarGz}); err != nil {
		return fmt.Errorf("failed to create final package tarball: %v", err)
	}

	if err := os.Chmod(filepath.Join(distPath, appPackage), 0666); err != nil {
		return fmt.Errorf("failed to set permissions for the final package: %v", err)
	}

	fmt.Printf("Rootfs package created: %s\n", filepath.Join(distPath, appPackage))
	return nil
}

func CompressRootfs(distPath, rootfsTarGzPath string) error {
	if err := fsutil.TarAndGzip(distPath, rootfsTarGzPath, []string{"."}); err != nil {
		return fmt.Errorf("failed to create rootfs.tar.gz: %v", err)
	}
	return nil
}

func CreateAndCompressRuntimeConfig(runtimeConfigPath string, appSettings configuration.AppSettings, buildMetadata map[string]string) error {
	if len(buildMetadata) > 0 && appSettings.Runtime.Annotations == nil {
		appSettings.Runtime.Annotations = make(map[string]string)
	}

	for key, value := range buildMetadata {
		appSettings.Runtime.Annotations[key] = value
	}
	configFilePath := filepath.Join(runtimeConfigPath, runtimeConfigFile)
	runtimeTarPath := filepath.Join(runtimeConfigPath, runtimeConfigTarFile)

	//Create Runtime config json
	if err := writeRuntimeSpecToJSON(appSettings, configFilePath); err != nil {
		return fmt.Errorf("failed to write runtime config: %v", err)
	}

	// Compress runtime config to tar
	if err := fsutil.CreateTarArchive(configFilePath, runtimeTarPath); err != nil {
		return fmt.Errorf("error creating runtime config tar archive: %v", err)
	}

	return nil
}
