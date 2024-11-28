// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package builder

import (
	"corteca/internal/configuration"
	"corteca/internal/fsutil"
	"corteca/internal/tui"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

const (
	buildArg                     = "--build-arg"
	runtimeConfigFile            = "config.json"
	adfFile                      = "ADF"
	nokiaRuntimeConfigAnnotation = "com.nokia.runtime.config"
)

var filesToRemove = []string{
	adfFile,
	runtimeConfigFile,
	"config.json.tar",
	"rootfs",
	"rootfs.tar.gz",
}

var InheritEnvironmentVars = []string{
	"HTTP_PROXY",
	"http_proxy",
	"HTTPS_PROXY",
	"https_proxy",
	"FTP_PROXY",
	"ftp_proxy",
	"NO_PROXY",
	"no_proxy",
	"MIRROR_REGISTRY",
}

func BuildContainer(imageName, arch, platform, appDir string, appSettings configuration.AppSettings, buildOptions configuration.BuildOptions) error {
	distPath := filepath.Join(appDir, "dist")
	outputType := strings.ToLower(buildOptions.OutputType)
	if err := fsutil.EnsureDirExists(distPath); err != nil {
		return fmt.Errorf("failed to create dist directory: %v", err)
	}

	args, err := prepareDockerBuildArgs(platform, buildOptions, appSettings, appDir, distPath, outputType)
	if err != nil {
		return err
	}

	if err := execDocker(args...); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	if err := handlePostBuildTasks(outputType, appDir, distPath, arch, platform, appSettings); err != nil {
		return err
	}

	return nil
}

func prepareDockerBuildArgs(platform string, buildOptions configuration.BuildOptions, appSettings configuration.AppSettings, appDir, distPath, outputType string) ([]string, error) {
	args := []string{"buildx", "build", "--platform", platform}

	if !buildOptions.SkipHostEnv {
		inheritEnvironment(&args)
	}

	if tui.DisableColoredOutput {
		args = append(args, "--progress=plain")
	}
	args = append(args, "-f", "Dockerfile")

	switch outputType {
	case "rootfs", "oci":
		rootfsPath := filepath.Join(distPath, "rootfs")
		if err := fsutil.CleanupOrCreateFolder(rootfsPath); err != nil {
			return nil, fmt.Errorf("failed to prepare dist directory: %v", err)
		}
		args = append(args, "--output", "type=local,dest="+rootfsPath)
	case "docker":
		args = append(args, "-t", fmt.Sprintf("%s:%s", appSettings.Name, appSettings.Version))
	default:
		return nil, fmt.Errorf("unknown image type: %s", buildOptions.OutputType)
	}

	args = append(args, appDir)
	return args, nil
}

func handlePostBuildTasks(outputType, appDir, distPath, arch, platform string, appSettings configuration.AppSettings) error {
	if outputType == "oci" {
		return handleOCITasks(distPath, arch, platform, appSettings)
	}

	if outputType == "rootfs" {
		return packageRootfs(appDir, distPath, arch, outputType, appSettings)
	}

	return nil
}

func handleOCITasks(distPath, arch, platform string, appSettings configuration.AppSettings) error {
	rootfsTarGz := filepath.Join(distPath, "rootfs.tar.gz")
	ociDirName := fmt.Sprintf("%s-%s-%s-oci", appSettings.Name, appSettings.Version, arch)
	ociTarName := fmt.Sprintf("%s-%s-%s-oci.tar", appSettings.Name, appSettings.Version, arch)
	ociDirPath := filepath.Join(distPath, ociDirName)
	ociTarPath := filepath.Join(distPath, ociTarName)

	if err := fsutil.TarAndGzip(distPath, rootfsTarGz, []string{"rootfs"}); err != nil {
		return fmt.Errorf("failed to create rootfs.tar.gz: %v", err)
	}

	if err := writeRuntimeSpecToJSON(appSettings, filepath.Join(distPath, runtimeConfigFile)); err != nil {
		return fmt.Errorf("failed to write runtime config: %v", err)
	}

	if err := CreateOCILayout(ociDirPath, rootfsTarGz, filepath.Join(distPath, runtimeConfigFile), arch, platform, appSettings); err != nil {
		return fmt.Errorf("failed to create OCI layout: %v", err)
	}

	if err := fsutil.CreateTarArchiveFromFolder(ociDirPath, ociTarPath); err != nil {
		return fmt.Errorf("failed to create OCI tar: %v", err)
	}

	filesToRemove = append(filesToRemove, ociDirName)
	if err := fsutil.RemoveFilesFromFolder(distPath, filesToRemove); err != nil {
		return fmt.Errorf("failed to clean up dist folder: %v", err)
	}

	return nil
}

func inheritEnvironment(args *[]string) {
	for _, varName := range InheritEnvironmentVars {
		varValue := os.Getenv(varName)
		if varValue != "" {
			(*args) = append(*args, buildArg, fmt.Sprintf("%s=%s", varName, varValue))
		}
	}
}

func EnableMultiplatformBuild(crossCompileCfg configuration.CrossCompileConfig) error {
	if !crossCompileCfg.Enabled {
		return nil // Cross-compilation not enabled
	}

	args := []string{"run", "--rm", "--privileged", crossCompileCfg.Image}
	args = append(args, crossCompileCfg.Args...)

	if err := execDocker(args...); err != nil {
		return fmt.Errorf("failed to setup cross-compilation: %v", err)
	}
	return nil
}

func execDocker(args ...string) error {
	// TODO: check if stdout is a tty first
	tui.SetOutputColor(tui.CBlue)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	tui.ResetOutputColor()
	return err
}

func packageRootfs(appDir, distPath, arch, outputType string, appSettings configuration.AppSettings) error {

	rootfsTarGz := "rootfs.tar.gz" // Name for the rootfs archive
	appPackage := fmt.Sprintf("%s-%s-%s-%s.tar.gz", appSettings.Name, appSettings.Version, arch, outputType)

	if err := fsutil.TarAndGzip(distPath, filepath.Join(distPath, rootfsTarGz), []string{"rootfs"}); err != nil {
		return fmt.Errorf("failed to create rootfs.tar.gz: %v", err)
	}

	_, err := fsutil.CopyFile(filepath.Join(appDir, adfFile), filepath.Join(distPath, adfFile))
	if err != nil {
		return fmt.Errorf("failed to copy ADF file: %v", err)
	}

	if err := fsutil.TarAndGzip(distPath, filepath.Join(distPath, appPackage), []string{adfFile, rootfsTarGz}); err != nil {
		return fmt.Errorf("failed to create final package tarball: %v", err)
	}

	if err := os.Chmod(filepath.Join(distPath, appPackage), 0666); err != nil {
		return fmt.Errorf("failed to set permissions for the final package: %v", err)
	}

	if err := fsutil.RemoveFilesFromFolder(distPath, filesToRemove); err != nil {
		return fmt.Errorf("failed to clean up dist folder: %v", err)
	}

	fmt.Printf("Rootfs package created: %s\n", filepath.Join(distPath, appPackage))
	return nil
}

func writeRuntimeSpecToJSON(appSettings configuration.AppSettings, filePath string) error {
	data, err := json.MarshalIndent(appSettings.Runtime, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal runtime spec: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}

func CreateOCILayout(ociDirPath, rootfsTarPath, runtimeConfigPath, arch, platform string, appSettings configuration.AppSettings) error {

	img := empty.Image

	rootfsLayer, err := tarball.LayerFromFile(rootfsTarPath)
	if err != nil {
		return fmt.Errorf("error creating rootfs layer: %v", err)
	}
	img, err = mutate.AppendLayers(img, rootfsLayer)
	if err != nil {
		return fmt.Errorf("error adding rootfs layer: %v", err)
	}

	runtimeTarPath := runtimeConfigPath + ".tar"
	if err := fsutil.CreateTarArchive(runtimeConfigPath, runtimeTarPath); err != nil {
		return fmt.Errorf("error creating runtime config tar archive: %v", err)
	}
	runtimeLayer, err := tarball.LayerFromFile(runtimeTarPath)
	if err != nil {
		return fmt.Errorf("error creating runtime config layer: %v", err)
	}
	runtimeLayerDigest, err := runtimeLayer.Digest()
	if err != nil {
		return fmt.Errorf("error creating runtime config layer digest: %v", err)
	}

	img, err = mutate.AppendLayers(img, runtimeLayer)

	if err != nil {
		return fmt.Errorf("error adding runtime config layer: %v", err)
	}

	cfg, err := img.ConfigFile()
	if err != nil {
		return fmt.Errorf("error getting image config file: %v", err)
	}
	cfg.Config.Entrypoint = []string{appSettings.Entrypoint}
	cfg.Config.Env = fsutil.ConvertMapToSlice(appSettings.Env)
	cfg.OS = strings.Split(platform, "/")[0]
	cfg.Architecture = strings.Split(platform, "/")[1]
	img, err = mutate.ConfigFile(img, cfg)
	if err != nil {
		return fmt.Errorf("error setting image config file: %v", err)
	}

	img = mutate.ConfigMediaType(img, types.OCIConfigJSON)
	img = mutate.MediaType(img, types.OCIManifestSchema1)
	img = mutate.Annotations(img, map[string]string{
		nokiaRuntimeConfigAnnotation: runtimeLayerDigest.String(),
	}).(v1.Image)

	ociPath, err := layout.Write(ociDirPath, empty.Index)
	if err != nil {
		return fmt.Errorf("error creating OCI layout: %v", err)
	}

	if err := ociPath.AppendImage(img); err != nil {
		return fmt.Errorf("error writing image to OCI layout: %v", err)
	}

	fmt.Println("Successfully created OCI image layout:", ociDirPath)
	return nil
}
