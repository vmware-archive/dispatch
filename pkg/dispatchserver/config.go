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
	DatabaseBackend  string `mapstructure:"db-backend" json:"db-backend"`
	DatabaseAddress  string `mapstructure:"db-file" json:"db-file"`
	DatabaseBucket   string `mapstructure:"db-database" json:"db-database"`
	DatabaseUsername string `mapstructure:"db-username" json:"db-username"`
	DatabasePassword string `mapstructure:"db-password" json:"db-password"`

	ResyncPeriod    time.Duration `mapstructure:"resync-period" json:"resync-period"`
	RegistryAuth    string        `mapstructure:"registry-auth" json:"registry-auth"`
	ImageRegistry   string        `mapstructure:"image-registry" json:"image-registry"`
	DisableRegistry bool          `mapstructure:"disable-registry" json:"disable-registry"`

	ImageManager    string `mapstructure:"image-manager" json:"image-manager"`
	FunctionManager string `mapstructure:"function-manager" json:"function-manager"`
	SecretsStore    string `mapstructure:"secret-store" json:"secret-store"`

	Host              string `mapstructure:"host" json:"host"`
	Port              int    `mapstructure:"port" json:"port"`
	DisableHTTP       bool   `mapstructure:"disable-http" json:"disable-http"`
	TLSPort           int    `mapstructure:"tls-port" json:"tls-port"`
	EnableTLS         bool   `mapstructure:"enable-tls" json:"enable-tls"`
	TLSCertificate    string `mapstructure:"tls-certificate" json:"tls-certificate"`
	TLSCertificateKey string `mapstructure:"tls-certificate-key" json:"tls-certificate-key"`
	LetsEncryptDomain string `mapstructure:"lets-encrypt-domain" json:"lets-encrypt-domain"`

	Tracer string `mapstructure:"tracer" json:"tracer"`
	Debug  bool   `mapstructure:"debug" json:"debug"`

	// Local server config options
	Local localConfig `mapstructure:"local" json:"local"`

	APIs apisConfig `mapstructure:"apis" json:"apis"`

	// Event Manager config options
	Events eventsConfig `mapstructure:"events" json:"events"`

	// Function Manager config options
	Functions functionsConfig `mapstructure:"functions" json:"functions"`

	// Identity Manager config options
	Identity identityConfig `mapstructure:"identity" json:"identity"`

	// Image Manager config options
	Image imageConfig `mapstructure:"image" json:"image"`
}

var defaultConfig = &serverConfig{}

func configGlobalFlags(flags *pflag.FlagSet) {
	flags.StringVar(&dispatchConfigPath, "config", "", "config file to use")

	flags.String("db-file", "./dispatch.db", "Database address, or database file path")
	flags.String("db-backend", "boltdb", "Database type to use")
	flags.String("db-database", "dispatch", "Database bucket or schema")
	flags.String("db-username", "dispatch", "Database username")
	flags.String("db-password", "dispatch", "Database password")

	flags.Duration("resync-period", 20*time.Second, "How often services should sync their state")
	flags.String("registry-auth", emptyRegistryAuth, "base64-encoded docker registry credentials")
	flags.String("image-registry", "dispatch", "Image registry host or docker hub org/username")
	flags.Bool("disable-registry", false, "Do not use image registry (do not push/pull images)")

	flags.String("image-manager", "", "URL to Image Manager")
	flags.String("function-manager", "", "URL to Function Manager")
	flags.String("secret-store", "", "URL to Secrets Store")

	flags.String("host", "127.0.0.1", "Host/IP to listen on")
	flags.Int("port", 8080, "HTTP port to listen on")
	flags.Bool("disable-http", false, "Disable HTTP Listener. TLS Listener must be enabled")
	flags.Int("tls-port", 8443, "TLS port to listen on")
	flags.String("tls-certificate", "", "Path to the certificate file")
	flags.String("tls-certificate-key", "", "Path to the certificate private key")
	flags.Bool("enable-tls", false, "Enable TLS (HTTPS) listener.")
	flags.String("lets-encrypt-domain", "", "Use Let's Encrypt certificate")

	flags.String("tracer", "", "OpenTracing-compatible Tracer URL")
	flags.Bool("debug", false, "Enable debugging logs")
}
