package webhook

import (
	"crypto/tls"
	"fmt"
)

// Config contains the server (the webhook) cert and key.
type config struct {
	CertFile string `json:"certFile,omitempty"`
	KeyFile  string `json:"keyFile,omitempty"`
	Addr     string `json:"port,omitempty"`
}

const (
	certFile string = "/tmp/k8s-webhook-server/serving-certs/tls.crt"
	keyFile  string = "/tmp/k8s-webhook-server/serving-certs/tls.key"
)

func configTLS(config *config) *tls.Config {
	config.CertFile = certFile
	config.KeyFile = keyFile
	config.Addr = "0.0.0.0:3443"
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		panic(fmt.Sprintf("load cert file error %v", err))
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}
}
