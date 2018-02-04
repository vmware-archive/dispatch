///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package config

// NO TESTS

import (
	"encoding/json"
	"fmt"
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
		Gateway string `json:"gateway"`
	} `json:"openfaas"`
	Riff struct {
		Gateway       string `json:"gateway"`
		K8sConfig     string `json:"k8s_config"`
		RiffNamespace string `json:"riff_namespace"`
	} `json:"riff"`
	Registry struct {
		RegistryURI  string `json:"uri"`
		RegistryAuth string `json:"auth"`
	} `json:"registry"`
}

// LoadConfiguration loads configurations from a local json file
func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err.Error())
	}
	return config
}
