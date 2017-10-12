///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package config

// NO TESTS

import (
	"encoding/json"
	"fmt"
	"os"
)

// Global contains global configuration variables
// and it is accessiable across the platfrom
var Global Config

// Config defines global configurations used through the serverless platfrom
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
		fmt.Println(err.Error())
	}
	return config
}
