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

// StaticUsers contains a list of static users
var StaticUsers Users

// Users defines a list of static users
type Users struct {
	Data []struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"users"`
}

// LoadStaticUsers loads static users from a local file
func LoadStaticUsers(file string) Users {
	var users Users
	usersFile, err := os.Open(file)
	defer usersFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(usersFile)
	jsonParser.Decode(&users)
	return users
}
