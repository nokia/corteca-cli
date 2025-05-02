package publish

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type artifactBlobHandler struct {
	artifactPath string
}

type readCloserWrapper struct {
	reader    io.Reader
	closeFunc func() error
}

func (r *readCloserWrapper) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *readCloserWrapper) Close() error {
	return r.closeFunc()
}

func (a *artifactBlobHandler) Get(ctx context.Context, repo string, hash v1.Hash) (io.ReadCloser, error) {
	reader, closer, err := findFileInArtifact(a.artifactPath, hash.String())
	if err != nil {
		return nil, err
	}

	return &readCloserWrapper{
		reader:    reader,
		closeFunc: closer,
	}, nil
}

func NewArtifactBlobHandler(artifact string) registry.BlobHandler {
	return &artifactBlobHandler{artifactPath: artifact}
}

func findFileInArtifact(artifactPath, targetName string) (io.Reader, func() error, error) {
	f, err := os.Open(artifactPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open archive: %w", err)
	}

	gzf, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	tarReader := tar.NewReader(gzf)

	// Extract the algorithmName/encoded portion from "<algorithm>:<encoded>" digest-format
	algorithmName, targetHash, err := splitDigest(targetName)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading tar file: %w", err)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error reading tar file: %w", err)
		}

		// Check if "blobs/<algorithm>/<encoded>" format matches header.Name
		if header.Name == filepath.ToSlash(filepath.Join("blobs", algorithmName, targetHash)) {
			return tarReader, func() error {
				errGZF := gzf.Close()
				errF := f.Close()

				if errGZF != nil {
					return errGZF
				}
				return errF
			}, nil
		}
	}

	gzf.Close()
	f.Close()
	return nil, nil, fmt.Errorf("file not found: %s", targetName)
}

func splitDigest(targetName string) (string, string, error) {
	parts := strings.SplitN(targetName, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid targetName format: expected <algorithm>:<hash>, got %s", targetName)
	}

	return parts[0], parts[1], nil
}

func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"test"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to marshal private key: %v", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load key pair: %v", err)
	}

	return cert, nil
}

func StartRegistry(address, artifact string) (*http.Server, error) {
	blobHandler := NewArtifactBlobHandler(artifact)

	handler := registry.New(registry.WithBlobHandler(blobHandler))

	cert, err := generateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed cert: %v", err)
	}

	server := &http.Server{
		Addr:    address,
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServeTLS(): %v", err)
		}
	}()

	time.Sleep(1 * time.Second)
	return server, nil
}
