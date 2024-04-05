// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package publish

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestHttpPut(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		fileContent    string
		token          string
		urlScheme      string
		expectedError  string
		httpStatusCode int
	}{
		{
			name:        "successful upload with token",
			fileContent: "some content",
			token:       "token",
			urlScheme:   "http",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			httpStatusCode: http.StatusOK,
		},
		{
			name:        "successful upload with basic auth",
			fileContent: "some content",
			token:       "",
			urlScheme:   "http",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}))
			},
			httpStatusCode: http.StatusCreated,
		},
		{
			name:          "unsupported URL scheme",
			fileContent:   "some content",
			token:         "token",
			urlScheme:     "ftp",
			expectedError: "unsupported format ftp",
		},
		{
			name:        "HTTP request fails with unauthorized",
			fileContent: "some content",
			token:       "token",
			urlScheme:   "http",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
				}))
			},
			expectedError: "server returned non-successful status: 401 Unauthorized",
		},
		{
			name:        "HTTP request fails with other error",
			fileContent: "some content",
			token:       "token",
			urlScheme:   "http",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectedError: "server returned non-successful status: 500 Internal Server Error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			tmpFile, err := os.CreateTemp("", "testfile-*.txt")
			if err != nil {
				t.Fatalf("Failed to create temporary file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.Write([]byte("test content"))
			if err != nil {
				t.Fatalf("Failed to write to temporary file: %v", tmpFile)
			}
			tmpFile.Close()

			var server *httptest.Server
			if tc.setupServer != nil {
				server = tc.setupServer()
				defer server.Close()
			}

			var testURL string
			if server != nil {
				testURL = server.URL
			}
			u, _ := url.Parse(testURL)
			u.Scheme = tc.urlScheme

			err = HttpPut(tmpFile.Name(), *u, tc.token)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
