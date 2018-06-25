///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package dispatchserver

import (
	"time"

	"github.com/spf13/pflag"
)

// emptyRegistryAuth == echo -n '{"username":"","password":"","email":""}' | base64
const emptyRegistryAuth = "eyJ1c2VybmFtZSI6IiIsInBhc3N3b3JkIjoiIiwiZW1haWwiOiIifQ=="

type serverConfig struct {
	// TODO: Refactor into Database connection string
	DatabaseBackend  string `mapstructure:"database-backend" json:"database-backend"`
	DatabaseAddress  string `mapstructure:"database-address" json:"database-address"`
	DatabaseBucket   string `mapstructure:"database-bucket" json:"database-bucket"`
	DatabaseUsername string `mapstructure:"database-username" json:"database-username"`
	DatabasePassword string `mapstructure:"database-password" json:"database-password"`

	ResyncPeriod  time.Duration `mapstructure:"resync-period" json:"resync-period"`
	RegistryAuth  string        `mapstructure:"registry-auth" json:"registry-auth"`
	ImageRegistry string        `mapstructure:"image-registry" json:"image-registry"`
	PushImages    bool          `mapstructure:"push-images" json:"push-images"`

	ImageManager    string `mapstructure:"image-manager" json:"image-manager"`
	FunctionManager string `mapstructure:"function-manager" json:"function-manager"`
	ServiceManager  string `mapstructure:"service-manager" json:"service-manager"`
	SecretsStore    string `mapstructure:"secrets-store" json:"secrets-store"`

	Host              string `mapstructure:"host" json:"host"`
	Port              int    `mapstructure:"port" json:"port"`
	DisableHTTP       bool   `mapstructure:"disable-http" json:"disable-http"`
	TLSPort           int    `mapstructure:"tls-port" json:"tls-port"`
	EnableTLS         bool   `mapstructure:"enable-tls" json:"enable-tls"`
	TLSCertificate    string `mapstructure:"tls-certificate" json:"tls-certificate"`
	TLSCertificateKey string `mapstrucutre:"tls-certificate-key" json:"tls-certificate-key"`

	Tracer string `mapstructure:"tracer" json:"tracer"`
	Debug  bool   `mapstructure:"debug" json:"debug"`

	Local localServer `mapstructure:"local" json:"local"`
}

var defaultConfig = &serverConfig{}

func configGlobalFlags(flags *pflag.FlagSet) {
	flags.StringVar(&dispatchConfigPath, "config", "", "config file to use")

	flags.String("database-address", "./dispatch.db", "Database address, or database file path")
	flags.String("database-backend", "boltdb", "Database type to use")
	flags.String("database-bucket", "dispatch", "Database bucket or schema")
	flags.String("database-username", "dispatch", "Database username")
	flags.String("database-password", "dispatch", "Database password")

	flags.Duration("resync-period", 10*time.Second, "How often services should sync their state")
	flags.String("registry-auth", emptyRegistryAuth, "base64-encoded docker registry credentials")
	flags.String("image-registry", "dispatch", "Image registry host or docker hub org/username")
	flags.Bool("push-images", false, "Push/pull images to/from image registry")

	flags.String("image-manager", "", "URL to Image Manager")
	flags.String("function-manager", "", "URL to Function Manager")
	flags.String("service-manager", "", "URL to Service Manager")
	flags.String("secrets-store", "", "URL to Secrets Store")

	flags.String("host", "127.0.0.1", "Host/IP to listen on")
	flags.Int("port", 8080, "HTTP port to listen on")
	flags.Bool("disable-http", false, "Disable HTTP Listener. TLS Listener must be enabled")
	flags.Int("tls-port", 8443, "TLS port to listen on")
	flags.String("tls-certificate", "", "Path to the certificate file")
	flags.String("tls-certificate-key", "", "Path to the certificate private key")
	flags.Bool("enable-tls", false, "Enable TLS (HTTPS) listener. tls-certificate and tls-certificate-key must be set")

	flags.String("tracer", "", "OpenTracing-compatible Tracer URL")
	flags.Bool("debug", false, "Enable debugging logs")
}
