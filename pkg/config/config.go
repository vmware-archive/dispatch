///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package config

// NO TESTS

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

// Global contains global configuration variables
var Global Config

// EmptyRegistryAuth == echo -n '{"username":"","password":"","email":""}' | base64
var EmptyRegistryAuth = "eyJ1c2VybmFtZSI6IiIsInBhc3N3b3JkIjoiIiwiZW1haWwiOiIifQ=="

// Config defines global configurations used in Dispatch
type Config struct {
	Identity struct {
		OIDCProvider string   `json:"oidc_provider"`
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURL  string   `json:"redirect_url"`
		Scopes       []string `json:"scopes"`
	} `json:"identity"`
	Openwhisk struct {
		AuthToken string `json:"auth_token"`
		Host      string `json:"host"`
	} `json:"openwhisk"`
	OpenFaas struct {
		Gateway       string `json:"gateway"`
		K8sConfig     string `json:"k8sConfig"`
		FuncNamespace string `json:"funcNamespace"`
	} `json:"openfaas"`
	Riff struct {
		Gateway       string `json:"gateway"`
		K8sConfig     string `json:"k8sConfig"`
		FuncNamespace string `json:"funcNamespace"`
	} `json:"riff"`
	Registry struct {
		RegistryURI  string `json:"uri"`
		RegistryAuth string `json:"auth"`
	} `json:"registry"`
}

// LoadConfiguration loads configurations from a local json file
func LoadConfiguration(file string) Config {
	configFile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func loadConfig(reader io.Reader) (Config, error) {
	var config Config
	jsonParser := json.NewDecoder(reader)
	err := jsonParser.Decode(&config)
	return config, err
}
