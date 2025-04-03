// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package packager

import (
	"corteca/internal/configuration"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

func writeBuildInfoJson(buildInfo map[string]string, rootfsPath string) error {

	data, err := json.Marshal(buildInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal buildinfo: %w", err)
	}

	file, err := os.Create(filepath.Join(rootfsPath, buildInfoPath))
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

func createOCILayout(ociDirPath, rootfsTarPath, runtimeConfigPath, platform string, appSettings configuration.AppSettings) error {
	img := empty.Image

	img, err := createRootfsLayer(img, rootfsTarPath)
	if err != nil {
		return err
	}

	img, runtimeLayerDigest, err := createRuntimeLayer(img, runtimeConfigPath)
	if err != nil {
		return err
	}

	img, err = updateImageConfig(img, appSettings, platform)
	if err != nil {
		return err
	}

	img = addAnnotationsToImage(img, runtimeLayerDigest)

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

func createRootfsLayer(img v1.Image, rootfsTarPath string) (v1.Image, error) {
	rootfsLayer, err := tarball.LayerFromFile(rootfsTarPath)
	if err != nil {
		return nil, fmt.Errorf("error creating rootfs layer: %v", err)
	}
	img, err = mutate.AppendLayers(img, rootfsLayer)
	if err != nil {
		return nil, fmt.Errorf("error adding rootfs layer: %v", err)
	}

	return img, nil
}

func createRuntimeLayer(img v1.Image, runtimeTarPath string) (v1.Image, string, error) {
	runtimeLayer, err := tarball.LayerFromFile(runtimeTarPath)
	if err != nil {
		return img, "", fmt.Errorf("error creating runtime config layer: %v", err)
	}
	runtimeLayerDigest, err := runtimeLayer.Digest()
	if err != nil {
		return img, "", fmt.Errorf("error creating runtime config layer digest: %v", err)
	}

	img, err = mutate.AppendLayers(img, runtimeLayer)
	if err != nil {
		return img, "", fmt.Errorf("error adding runtime config layer: %v", err)
	}
	return img, runtimeLayerDigest.String(), nil
}

func updateImageConfig(img v1.Image, appSettings configuration.AppSettings, platform string) (v1.Image, error) {
	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("error getting image config file: %v", err)
	}
	// sets entrypoint, env, os and architecture
	cfg.Config.Entrypoint = []string{appSettings.Entrypoint}
	cfg.Config.Env = make([]string, 0, len(appSettings.Env))
	for key, value := range appSettings.Env {
		cfg.Config.Env = append(cfg.Config.Env, fmt.Sprintf("%s=%s", key, value))
	}
	cfg.OS = strings.Split(platform, "/")[0]
	cfg.Architecture = strings.Split(platform, "/")[1]
	img, err = mutate.ConfigFile(img, cfg)
	if err != nil {
		return nil, fmt.Errorf("error setting image config file: %v", err)
	}

	// setting media type of the image and annotations
	img = mutate.ConfigMediaType(img, types.OCIConfigJSON)
	// image's format
	img = mutate.MediaType(img, types.OCIManifestSchema1)

	return img, nil
}

func addAnnotationsToImage(img v1.Image, runtimeLayerDigest string) v1.Image {
	img = mutate.Annotations(img, map[string]string{
		nokiaRuntimeConfigAnnotation: runtimeLayerDigest,
	}).(v1.Image)

	return img
}
