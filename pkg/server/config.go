///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"github.com/spf13/pflag"
)

type serverConfig struct {
	ImageRegistry string `mapstructure:"image-registry" json:"image-registry"`

	K8sConfig string `mapstructure:"kubeconfig" json:"kubeconfig,omitempty"`
	// Each org should have their own namespace, but we don't have org support yet.
	// The org should also creat the /store volume (in that namespace).
	Namespace  string `mapstructure:"namespace" json:"namespace"`
	SourceRoot string `mapstructure:"sourceroot" json:"sourceroot,omitempty"`

	Host              string `mapstructure:"host" json:"host"`
	Port              int    `mapstructure:"port" json:"port"`
	DisableHTTP       bool   `mapstructure:"disable-http" json:"disable-http"`
	TLSPort           int    `mapstructure:"tls-port" json:"tls-port"`
	EnableTLS         bool   `mapstructure:"enable-tls" json:"enable-tls"`
	TLSCertificate    string `mapstructure:"tls-certificate" json:"tls-certificate"`
	TLSCertificateKey string `mapstrucutre:"tls-certificate-key" json:"tls-certificate-key"`

	Debug bool `mapstructure:"debug" json:"debug"`
}

var defaultConfig = &serverConfig{}

func configGlobalFlags(flags *pflag.FlagSet) {
	flags.StringVar(&dispatchConfigPath, "config", "", "config file to use")

	flags.String("image-registry", "dispatch", "Image registry host or docker hub org/username")

	flags.String("kubeconfig", "", "Path to kubeconfig")
	flags.String("namespace", "", "Dispatch namespace")
	flags.String("sourceroot", "", "Path source storage volume")

	flags.String("host", "127.0.0.1", "Host/IP to listen on")
	flags.Int("port", 8080, "HTTP port to listen on")
	flags.Bool("disable-http", false, "Disable HTTP Listener. TLS Listener must be enabled")
	flags.Int("tls-port", 8443, "TLS port to listen on")
	flags.String("tls-certificate", "", "Path to the certificate file")
	flags.String("tls-certificate-key", "", "Path to the certificate private key")
	flags.Bool("enable-tls", false, "Enable TLS (HTTPS) listener.")

	flags.Bool("debug", false, "Enable debugging logs")
}
