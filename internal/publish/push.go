package publish

import (
	"corteca/internal/configuration"
	"corteca/internal/fsutil"
	"corteca/internal/tui"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pterm/pterm"
)

func PushImage(imagePath string, registryURL *url.URL, token, tag string, withProgress bool) error {
	distDir := filepath.Dir(imagePath)
	extractedImagePath := strings.TrimSuffix(imagePath, ".tar")
	extractedOCIName := filepath.Base(extractedImagePath)

	if err := fsutil.ExtractTarball(imagePath, extractedImagePath); err != nil {
		return fmt.Errorf("failed to extract OCI image: %w", err)
	}

	versionRef, err := name.NewTag(fmt.Sprintf("%s%s/%s", registryURL.Host, registryURL.Path, tag))
	if err != nil {
		return fmt.Errorf("failed to parse image reference: %w", err)
	}

	index, err := layout.ImageIndexFromPath(extractedImagePath)
	if err != nil {
		return fmt.Errorf("failed to read image index from path: %w", err)
	}

	manifest, err := index.IndexManifest()
	if err != nil {
		return fmt.Errorf("failed to get index manifest: %w", err)
	}

	var img v1.Image
	for _, desc := range manifest.Manifests {
		image, err := index.Image(desc.Digest)
		if err != nil {
			return fmt.Errorf("failed to get image: %w", err)
		}
		img = image
		break
	}

	transport := remote.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})

	auth, err := getAuthenticator(registryURL, token)
	if err != nil {
		return fmt.Errorf("failed to get authenticator: %w", err)
	}

	options := []remote.Option{
		remote.WithAuth(auth),
		transport,
	}

	if withProgress {
		updates := make(chan v1.Update, 8)
		progressBar := initializeProgressBar()
		go handleProgressUpdates(progressBar, updates)
		options = append(options, remote.WithProgress(updates))
	}

	if err := remote.Write(versionRef, img, options...); err != nil {
		return fmt.Errorf("failed to push image manifest to registry: %w", err)
	}

	if err := fsutil.RemoveFilesFromFolder(distDir, []string{extractedOCIName}); err != nil {
		return fmt.Errorf("failed to clean up extracted files: %w", err)
	}
	tui.DisplaySuccessMsg(fmt.Sprintf("Pushed image '%v' as '%v'\n", imagePath, versionRef.Name()))
	return nil
}

func handleProgressUpdates(bar *pterm.ProgressbarPrinter, updates chan v1.Update) {
	var lastComplete int64
	var totalSizeSet bool
	for update := range updates {
		if !totalSizeSet && update.Total > 0 {
			bar.Total = int(update.Total)
			totalSizeSet = true
		}
		progress := int(update.Complete - lastComplete)
		bar.Add(progress)
		lastComplete = update.Complete
		//pterm.Debug.Println(fmt.Sprintf("Progress: %d/%d", update.Complete, update.Total))
	}
	bar.Stop()
}

func getAuthenticator(registryURL *url.URL, token string) (authn.Authenticator, error) {
	if token != "" {
		return &authn.Bearer{
			Token: token,
		}, nil
	} else {
		// registryURL should always include a valid credentials or authentication token
		password, _ := registryURL.User.Password()
		return authn.FromConfig(authn.AuthConfig{
			Username: registryURL.User.Username(),
			Password: password,
		}), nil
	}
}

func initializeProgressBar() *pterm.ProgressbarPrinter {
	bar, _ := pterm.DefaultProgressbar.WithTotal(100).WithTitle("Pushing").Start()
	return bar
}

func GenerateTag(appSettings configuration.AppSettings, filePath, arch string) (string, error) {

	appName := appSettings.Name
	appVersion := appSettings.Version
	return fmt.Sprintf("%s/%s:%s", arch, appName, appVersion), nil
}
