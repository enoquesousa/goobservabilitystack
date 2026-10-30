package tracing_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type traceIdentity struct {
	certificate tls.Certificate
	certPath    string
	keyPath     string
}

type traceAuthority struct {
	certificate *x509.Certificate
	key         *ecdsa.PrivateKey
	path        string
	roots       *x509.CertPool
}

type tracePKI struct {
	authority traceAuthority
	server    traceIdentity
	client    traceIdentity
}

func newTraceAuthority(t *testing.T) traceAuthority {
	t.Helper()
	key := newTraceKey(t)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "synthetic-tracing-root"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		BasicConstraintsValid: true, IsCA: true, KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certificate, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "root.crt")
	writeTracePEM(t, path, "CERTIFICATE", der)
	roots := x509.NewCertPool()
	roots.AddCert(certificate)
	return traceAuthority{certificate: certificate, key: key, path: path, roots: roots}
}

func newTraceKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func issueTraceIdentity(t *testing.T, authority traceAuthority, template *x509.Certificate) traceIdentity {
	t.Helper()
	key := newTraceKey(t)
	template.SerialNumber = big.NewInt(2)
	template.NotBefore, template.NotAfter = time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	template.KeyUsage = x509.KeyUsageDigitalSignature
	der, err := x509.CreateCertificate(rand.Reader, template, authority.certificate, &key.PublicKey, authority.key)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath, keyPath := filepath.Join(dir, "identity.crt"), filepath.Join(dir, "identity.key")
	writeTracePEM(t, certPath, "CERTIFICATE", der)
	writeTracePEM(t, keyPath, "PRIVATE KEY", privateKey)
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
	return traceIdentity{certificate: certificate, certPath: certPath, keyPath: keyPath}
}

func writeTracePEM(t *testing.T, path, blockType string, contents []byte) {
	t.Helper()
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: contents}), 0o600); err != nil {
		t.Fatal(err)
	}
}

func newTracePKI(t *testing.T) tracePKI {
	t.Helper()
	authority := newTraceAuthority(t)
	server := issueTraceIdentity(t, authority, &x509.Certificate{
		Subject:     pkix.Name{CommonName: "synthetic-trace-collector"},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	})
	client := issueTraceIdentity(t, authority, &x509.Certificate{
		Subject:     pkix.Name{CommonName: "synthetic-trace-client"},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	})
	return tracePKI{authority: authority, server: server, client: client}
}

func (pki tracePKI) collectorTLS(requireClient bool) *tls.Config {
	config := &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{pki.server.certificate}}
	if requireClient {
		config.ClientAuth, config.ClientCAs = tls.RequireAndVerifyClientCert, pki.authority.roots
	}
	return config
}
