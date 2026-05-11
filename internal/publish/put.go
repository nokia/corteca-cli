// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package publish

import (
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/tui"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const (
	authHttpBasicName  = "basic"
	authHttpBearerName = "bearer"
	authHttpDigestName = "digest"
)

type PutReader struct {
	file  afero.File
	ch    chan<- tui.ProgressUpdate
	total int64
}

func (r *PutReader) Read(p []byte) (int, error) {
	n, err := r.file.Read(p)
	if err != nil {
		return n, err
	}
	pos, err := r.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return n, err
	}
	if r.total == 0 {
		if r.total, err = r.file.Seek(0, io.SeekEnd); err != nil {
			return n, err
		}
		if _, err = r.file.Seek(pos, io.SeekStart); err != nil {
			return n, err
		}
	}
	r.ch <- tui.ProgressUpdate{
		Current: pos,
		Total:   r.total,
	}
	return n, err
}

func HttpPut(filePath string, url url.URL, token string) error {

	if url.Scheme != "http" && url.Scheme != "https" {
		return fmt.Errorf("unsupported format %s", url.Scheme)
	}

	fs := afero.NewOsFs()
	fileName := filepath.Base(filePath)
	url.Path = filepath.Join(url.Path, fileName)

	file, err := fs.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	prog := tui.PromptForProgress(fmt.Sprintf("Uploading %s", fileName))
	defer close(prog)

	req, err := http.NewRequest("PUT", url.String(), &PutReader{file: file, ch: prog})
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)

	} else {
		// Require url to contain a password
		password, _ := url.User.Password()
		req.SetBasicAuth(url.User.Username(), password)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned non-successful status: %s", resp.Status)
	}

	tui.DisplaySuccessMsg(fmt.Sprintf("Successfully uploaded file '%v' to '%v'", fileName, url.Redacted()))

	return nil
}

func AuthenticateHttp(config configuration.HttpClientEndpoint) (*url.URL, error) {

	u, err := url.Parse(config.Addr.String())
	if err != nil {
		return nil, err
	}

	authType := strings.ToLower(config.Auth)
	switch authType {
	case authHttpBasicName:

		username := config.Username.String()
		password := config.Password.String()

		// Check for username in .yaml config
		if username == "" {
			// Prompt for username
			username, err = tui.PromptForValue("Enter username", "")
			if err != nil {
				return nil, err
			}
		}

		// Check for password in config
		if password == "" {
			// Prompt for password
			password, err = tui.PromptForPassword("Enter password")
			if err != nil {
				return nil, err

			}
		}

		u.User = url.UserPassword(username, password)

	case authHttpBearerName:
		if config.Token.String() == "" {
			return nil, errors.New("no bearer token present in configuration even though HTTP Bearer authentication has been requested")
		}
	case authHttpDigestName:
		// TODO: implement
		return nil, errors.New("digest HTTP authentication not implemented yet")
	}

	return u, nil
}
