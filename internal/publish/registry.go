// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package publish

import (
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/tui"
	"net/http"

	"github.com/google/go-containerregistry/pkg/registry"
)

func StartRegistry(config configuration.HttpServerEndpoint) (*http.Server, error) {
	handler := registry.New()
	server := &http.Server{Addr: config.Addr.String(), Handler: handler}
	certFile := config.Certificate.String()
	keyFile := config.Key.String()
	go func() {
		var err error
		if len(certFile) > 0 {
			err = server.ListenAndServeTLS(certFile, keyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			tui.LogError("Error while running registry server: %s", err.Error())
		}
	}()
	return server, nil
}
