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
		ImageRegistry string `json:"image_registry"`
		RegistryAuth  string `json:"registry_auth"`
	} `json:"openfaas"`
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
