// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package publish

import (
	"corteca/internal/configuration"
	"corteca/internal/tui"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/spf13/afero"
)

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

	progressReader, err := tui.PromptForProgress(file, fmt.Sprintf("Uploading %s", fileName))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url.String(), progressReader)
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
	progressReader.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned non-successful status: %s", resp.Status)
	}

	tui.DisplaySuccessMsg(fmt.Sprintf("Successfully uploaded file '%v' to '%v'", fileName, url.Redacted()))

	return nil
}

func AuthenticateHttp(target configuration.PublishTarget) (*url.URL, string, error) {
	var token string

	u, err := url.Parse(target.Addr)
	if err != nil {
		return nil, "", err
	}

	authType := target.Auth

	switch authType {
	case configuration.AUTH_HTTP_BASIC:

		username := u.User.Username()
		if username == "" {
			username, err = tui.PromptForValue("Enter username", "")
			if err != nil {
				return nil, "", err
			}
		}

		password, ok := u.User.Password()
		if !ok {
			password, err = tui.PromptForPassword("Enter password")
			if err != nil {
				return nil, "", err
			}
		}

		u.User = url.UserPassword(username, password)

	case configuration.AUTH_HTTP_BEARER:
		if token = target.Token; token == "" {
			return nil, "", errors.New("no bearer token present in configuration even though HTTP Bearer authentication has been requested")
		}
	case configuration.AUTH_HTTP_DIGEST:
		// TODO: implement
		return nil, "", errors.New("digest HTTP authentication not implemented yet")
	}

	return u, token, nil
}
