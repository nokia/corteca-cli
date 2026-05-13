// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package configuration

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/icholy/digest"
)

const (
	BasicClientAuth  = "basic"
	BearerClientAuth = "bearer"
	DigestClientAuth = "digest"
)

type HttpServerEndpoint struct {
	Endpoint    `yaml:",inline"`
	Certificate TemplateField `yaml:"certificate"`
	Key         TemplateField `yaml:"key"`
}

type HttpClientEndpoint struct {
	Endpoint            `yaml:",inline"`
	Auth                string        `yaml:"auth,omitempty"`
	Username            TemplateField `yaml:"username,omitempty"`
	Password            TemplateField `yaml:"password,omitempty"`
	Token               TemplateField `yaml:"token,omitempty"`
	SkipTLSVerification bool          `yaml:"skipTLSVerification"`
}

// transport to use basic authentication
type BasicAuthTransport struct {
	Username  string
	Password  string
	Transport http.RoundTripper
}

func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.SetBasicAuth(t.Username, t.Password)
	return t.transport().RoundTrip(req2)
}

func (t *BasicAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// transport to use bearer authentication
type BearerAuthTransport struct {
	Token     string
	Transport http.RoundTripper
}

func (t *BearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", "Bearer "+t.Token)
	return t.transport().RoundTrip(req2)
}

func (t *BearerAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport

}
func (ep *HttpClientEndpoint) NewHttpClient() (*http.Client, error) {
	token := ep.Token.String()
	username := ep.Username.String()
	password := ep.Password.String()
	var bearer, basic http.RoundTripper
	if len(token) > 0 {
		bearer = &BearerAuthTransport{Token: token}
	}
	if len(username) > 0 || len(password) > 0 {
		basic = &BasicAuthTransport{Username: username, Password: password}
	}
	client := &http.Client{}
	switch strings.ToLower(ep.Auth) {
	case BasicClientAuth:
		client.Transport = basic
	case BearerClientAuth:
		client.Transport = bearer
	case DigestClientAuth:
		client.Transport = &digest.Transport{Username: username, Password: password}
	case "":
		// if no explicit auth method specified, prioritize bearer
		if bearer != nil {
			client.Transport = bearer
		} else if basic != nil {
			client.Transport = basic
		}
	default:
		return nil, fmt.Errorf("unknown HTTP authentication '%s'", ep.Auth)
	}
	return client, nil
}

type SSHClientEndpoint struct {
	Endpoint       `yaml:",inline"`
	Username       TemplateField `yaml:"username,omitempty"`
	Password       TemplateField `yaml:"password,omitempty"`
	Password2      TemplateField `yaml:"password2,omitempty"`
	PrivateKeyFile TemplateField `yaml:"privateKeyFile,omitempty"`
}
