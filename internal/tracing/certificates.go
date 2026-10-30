package tracing

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

func validateCertificate(name string) error {
	path := setting(name)
	if path == "" {
		return nil
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return invalidSetting(name, "must identify a readable PEM trust certificate file")
	}
	if !x509.NewCertPool().AppendCertsFromPEM(contents) {
		return invalidSetting(name, "must contain a valid PEM trust certificate")
	}
	return nil
}

func validateClientCertificate(prefix string) error {
	certName, keyName := prefix+"CLIENT_CERTIFICATE", prefix+"CLIENT_KEY"
	certPath, keyPath := setting(certName), setting(keyName)
	if certPath == "" && keyPath == "" {
		return nil
	}
	if certPath == "" || keyPath == "" {
		return invalidSetting(certName+" and "+keyName, "must be configured together")
	}
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		return invalidSetting(certName+" and "+keyName, "must identify a readable, matching PEM certificate/key pair")
	}
	return nil
}
