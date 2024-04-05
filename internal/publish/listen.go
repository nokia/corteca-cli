// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package publish

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func ListenAsync(serverRoot string, addr *url.URL) (*http.Server, error) {
	// obtain protocol (http, if none specified)
	protocol := strings.ToLower(addr.Scheme)
	if protocol == "" {
		protocol = "http"
	}

	switch protocol {
	case "http":
		l, err := net.Listen("tcp", fmt.Sprintf("%v:%v", addr.Hostname(), addr.Port()))
		if err != nil {
			return nil, err
		}
		srv := &http.Server{
			Addr:    l.Addr().String(),
			Handler: http.FileServer(http.Dir(serverRoot)),
		}
		go func() {
			srv.Serve(l)
		}()
		return srv, nil
	default:
		return nil, fmt.Errorf("unsupported protocol '%v'", protocol)
	}
}
