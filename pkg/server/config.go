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
	Namespace string `mapstructure:"namespace" json:"namespace"`
	// Storage config for function storage [minio (s3)/file (nfs)]
	Storage        string `mapstructure:"storage" json:"storage,omitempty"`
	FileSourceRoot string `mapstructure:"file-sourceroot" json:"file-sourceroot,omitempty"`
	MinioUsername  string `mapstructure:"minio-username" json:"minio-username,omitempty"`
	MinioPassword  string `mapstructure:"minio-password" json:"minio-password,omitempty"`
	MinioAddress   string `mapstructure:"minio-address" json:"minio-address,omitempty"`

	BuildImage       string `mapstructure:"build-image" json:"build-image"`
	IngressGatewayIP string `mapstructure:"ingress-gateway-ip" json:"ingress-gateway-ip"`
	InternalGateway  string `mapstructure:"internal-gateway" json:"internal-gateway"`
	SharedGateway    string `mapstructure:"shared-gateway" json:"shared-gateway"`
	DispatchHost     string `mapstructure:"dispatch-host" json:"dispatch-host"`

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

	flags.String("storage", "minio", "Function storage [minio|file]")
	flags.String("file-sourceroot", "/store", "Path source storage volume")
	flags.String("minio-username", "", "Minio username")
	flags.String("minio-password", "", "Minio password")
	flags.String("minio-address", "minio.minio.svc.cluster.local:9000", "Minio address (host:port)")

	flags.String("build-image", defaultBuildImage, "Docker image for building functions")
	flags.String("ingress-gateway-ip", "", "IP of knative ingress gateway (default is empty)")
	flags.String("internal-gateway", "knative-ingressgateway.istio-system.svc.cluster.local", "Knative/Istio internal gateway")
	flags.String("shared-gateway", "knative-shared-gateway.knative-serving.svc.cluster.local", "Knative/Istio shared gateway")
	flags.String("dispatch-host", "dispatch.local", "Dispatch host DNS name")

	flags.String("host", "127.0.0.1", "Host/IP to listen on")
	flags.Int("port", 8080, "HTTP port to listen on")
	flags.Bool("disable-http", false, "Disable HTTP Listener. TLS Listener must be enabled")
	flags.Int("tls-port", 8443, "TLS port to listen on")
	flags.String("tls-certificate", "", "Path to the certificate file")
	flags.String("tls-certificate-key", "", "Path to the certificate private key")
	flags.Bool("enable-tls", false, "Enable TLS (HTTPS) listener.")

	flags.Bool("debug", false, "Enable debugging logs")
}
