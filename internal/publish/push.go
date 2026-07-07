// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package publish

import (
	"compress/gzip"
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/fsutil"
	"github.com/nokia/corteca-cli/internal/tui"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

type gzipReadCloser struct {
	*gzip.Reader
	file *os.File
}

func GzipOpener(path string) tarball.Opener {
	return func() (io.ReadCloser, error) {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		gz, err := gzip.NewReader(f)
		if err != nil {
			_ = f.Close()
			return nil, err
		}

		// combine both closers
		return &gzipReadCloser{
			Reader: gz,
			file:   f,
		}, nil
	}
}

func (g *gzipReadCloser) Close() error {
	if err := g.Reader.Close(); err != nil {
		return err
	}
	return g.file.Close()
}

func PushImage(tarballPath string, target *configuration.HttpClientEndpoint, withProgress bool) error {
	// create image tag from URL
	url, err := url.Parse(target.Addr.String())
	if err != nil {
		return err
	}
	tagOpts := []name.Option{name.StrictValidation}
	if url.Scheme == "http" {
		tagOpts = append(tagOpts, name.Insecure)
	}
	tag, err := name.NewTag(strings.ToLower(url.Host+url.Path), tagOpts...)
	if err != nil {
		return err
	}

	// get image from tarball
	tmp, err := os.MkdirTemp("", "corteca_image_")
	if err != nil {
		return fmt.Errorf("cannot create tmp folder: %w", err)
	}
	if err := fsutil.ExtractTarball(tarballPath, tmp); err != nil {
		return fmt.Errorf("cannot extract %s to %s: %w", tarballPath, tmp, err)
	}
	lp, err := layout.FromPath(tmp)
	if err != nil {
		return fmt.Errorf("cannot open OCI layout from %s: %w", tmp, err)
	}
	idx, err := lp.ImageIndex()
	if err != nil {
		return fmt.Errorf("cannot open index from %s: %w", tmp, err)
	}
	im, err := idx.IndexManifest()
	if err != nil {
		return fmt.Errorf("cannot open index manifest from %s: %w", tmp, err)
	}
	image, err := idx.Image(im.Manifests[0].Digest)
	if err != nil {
		return fmt.Errorf("cannot open image from %s: %w", tmp, err)
	}

	// set client options
	clientOpts := []remote.Option{}
	if target.SkipTLSVerification {
		clientOpts = append(clientOpts, remote.WithTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}))
	}
	switch target.Auth {
	case configuration.BasicClientAuth:
		clientOpts = append(clientOpts, remote.WithAuth(authn.FromConfig(authn.AuthConfig{Username: target.Username.String(), Password: target.Password.String()})))
	case configuration.BearerClientAuth:
		clientOpts = append(clientOpts, remote.WithAuth(authn.FromConfig(authn.AuthConfig{RegistryToken: target.Token.String()})))
	}
	if withProgress {
		updates := make(chan v1.Update, 8)
		prog := tui.PromptForProgress(fmt.Sprintf("Pushing %s", tag.String()))
		defer close(prog)
		clientOpts = append(clientOpts, remote.WithProgress(updates))
		go func() {
			for update := range updates {
				prog <- tui.ProgressUpdate{Current: update.Complete, Total: update.Total}
			}
		}()
	}

	if err := remote.Write(tag, image, clientOpts...); err != nil {
		return fmt.Errorf("failed to push image manifest to registry: %w", err)
	}
	tui.DisplaySuccessMsg(fmt.Sprintf("Pushed image %s to %s", tarballPath, tag.String()))
	return nil
}
