// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package builder

import (
	"corteca/internal/configuration"
	"corteca/internal/fsutil"
	"corteca/internal/templating"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	csReset    = "\033[0m"
	sBold      = "\033[1m"
	sUnderline = "\033[4m"
	sStrike    = "\033[9m"
	sItalic    = "\033[3m"

	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cBlue   = "\033[34m"
	cPurple = "\033[35m"
	cCyan   = "\033[36m"
	cWhite  = "\033[37m"
)

const buildArg = "--build-arg"

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

var (
	distPath   string
	rootfsPath string
	ociPath    string
	outputType string
)

func BuildContainer(imageName string, arch string, platform string, appDir string, appName string, appVersion string,
	buildOptions configuration.BuildOptions) error {
	distPath = filepath.Join(appDir, "dist")
	outputType = strings.ToLower(buildOptions.OutputType)

	args := []string{
		"buildx", "build",
		"--platform", platform,
	}

	if !buildOptions.SkipHostEnv {
		inheritEnvironment(&args)
	}
	args = append(args, "-f", "Dockerfile")

	switch outputType {
	case "rootfs":
		rootfsPath = filepath.Join(distPath, "rootfs")
		if err := fsutil.CleanupOrCreateFolder(rootfsPath); err != nil {
			return fmt.Errorf("failed to prepare dist directory: %v", err)
		}
		args = append(args, "--output", "type=local,dest="+rootfsPath)
	case "docker":
		args = append(args, "-t", fmt.Sprintf("%s:%s", appName, appVersion))
	case "oci":
		ociPath = filepath.Join(distPath, "oci-images")
		if err := fsutil.CleanupOrCreateFolder(ociPath); err != nil {
			return fmt.Errorf("failed to prepare OCI directory: %v", err)
		}
		ociTarName := fmt.Sprintf("%s-%s.tar", appName, appVersion)
		args = append(args, "--output", "type=oci,oci-mediatypes=true,dest="+filepath.Join(ociPath, ociTarName))
	default:
		return fmt.Errorf("unknown image type: %s", buildOptions.OutputType)
	}
	args = append(args, appDir)

	if err := execDocker(args...); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	if outputType == "rootfs" {
		if err := packageRootfs(appDir, appName, appVersion, arch); err != nil {
			return fmt.Errorf("failed to package rootfs: %v", err)
		}
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
	fmt.Fprint(os.Stdout, cBlue)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Fprint(os.Stdout, csReset)
	return err
}

func packageRootfs(appDir, appName, appVersion, arch string) error {
	adfPath := "ADF"               // Relative path to ADF within appDir
	rootfsTarGz := "rootfs.tar.gz" // Name for the rootfs archive
	appPackage := fmt.Sprintf("%s-%s-%s-%s.tar.gz", appName, appVersion, arch, "rootfs")

	if err := fsutil.TarAndGzip(distPath, filepath.Join(distPath, rootfsTarGz), []string{"rootfs"}); err != nil {
		return fmt.Errorf("failed to create rootfs.tar.gz: %v", err)
	}

	_, err := fsutil.CopyFile(filepath.Join(appDir, adfPath), filepath.Join(distPath, adfPath))
	if err != nil {
		return fmt.Errorf("failed to copy ADF file: %v", err)
	}

	if err := fsutil.TarAndGzip(distPath, filepath.Join(distPath, appPackage), []string{adfPath, rootfsTarGz}); err != nil {
		return fmt.Errorf("failed to create final package tarball: %v", err)
	}

	if err := os.Chmod(filepath.Join(distPath, appPackage), 0666); err != nil {
		return fmt.Errorf("failed to set permissions for the final package: %v", err)
	}

	fmt.Printf("Rootfs package created: %s\n", filepath.Join(distPath, appPackage))
	return nil
}

func GenerateDockerfileFromTemplate(context map[string]interface{}) error {
	buildContext, ok := context["build"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("build settings not found in context")
	}
	templateContent, ok := buildContext["dockerFileTemplate"].(string)
	if !ok {
		return fmt.Errorf("dockerFileTemplate not found or is not a string")
	}

	dockerfileContent, err := templating.RenderTemplateString(templateContent, context)
	if err != nil {
		return fmt.Errorf("error executing Dockerfile template: %w", err)
	}

	if err := os.WriteFile("Dockerfile", []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("error writing Dockerfile: %w", err)
	}

	return nil
}
