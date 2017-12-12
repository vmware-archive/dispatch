///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/testing/dev"
)

func getTestKongInstance(t *testing.T) *Client {
	log.SetLevel(log.DebugLevel)

	c, err := NewClient(&Config{
		Host:     "http://192.168.99.100:8002",
		Upstream: "function-manager",
	})
	assert.Nil(t, err)

	err = c.Initialize()
	assert.Nil(t, err)
	return c
}

func assertAPIEqual(t *testing.T, expected *gateway.API, real *gateway.API) {

	assert.Equal(t, expected.Name, real.Name)
	assert.Equal(t, expected.Hosts, real.Hosts)
	assert.Equal(t, expected.Methods, real.Methods)
	assert.Equal(t, expected.URIs, real.URIs)
}

func TestAddAPI(t *testing.T) {

	dev.EnsureLocal(t)
	client := getTestKongInstance(t)

	expected := &gateway.API{
		Name:           "testAddAPI",
		Function:       "testFunction",
		Authentication: "public",
		Enabled:        true,
		Hosts:          []string{"test.com", "vmware.com"},
		URIs:           []string{"/test", "/hello"},
		Methods:        []string{"GET", "POST"},
		Protocols:      []string{"http", "https"},
		TLS:            "testtls",
	}

	// clear
	client.DeleteAPI(expected)

	real, err := client.AddAPI(expected)
	assert.Nil(t, err)
	assertAPIEqual(t, expected, real)

	err = client.DeleteAPI(expected)
	assert.Nil(t, err)
}

func TestUpdateAPIWithCors(t *testing.T) {

	dev.EnsureLocal(t)
	client := getTestKongInstance(t)

	expected := &gateway.API{
		Name:           "testUpdateAPIWithCORS",
		Function:       "testFunction",
		Authentication: "public",
		Enabled:        true,
		Hosts:          []string{"test.com", "vmware.com"},
		URIs:           []string{"/test", "/hello"},
		Methods:        []string{"GET", "POST"},
		Protocols:      []string{"http", "https"},
		TLS:            "testtls",
		CORS:           true,
	}

	// clear
	client.DeleteAPI(expected)

	real, err := client.UpdateAPI(expected.Name, expected)
	assert.Nil(t, err)
	assertAPIEqual(t, expected, real)

	err = client.DeleteAPI(expected)
	assert.Nil(t, err)
}

func TestUpdateAPI(t *testing.T) {

	dev.EnsureLocal(t)
	client := getTestKongInstance(t)
	expected := &gateway.API{
		Name:           "testUpdateAPI",
		Function:       "testFunction",
		Authentication: "public",
		Enabled:        true,
		Hosts:          []string{"test.com", "vmware.com"},
		URIs:           []string{"/test", "/hello"},
		Methods:        []string{"GET", "POST"},
		Protocols:      []string{"http", "https"},
		TLS:            "testtls",
	}

	// clear
	client.DeleteAPI(expected)

	real, err := client.AddAPI(expected)
	assert.Nil(t, err)
	assertAPIEqual(t, expected, real)

	expected.ID = real.ID
	expected.CreatedAt = real.CreatedAt
	expected.Methods = []string{"PATCH"}
	expected.Hosts = []string{"updated.com", "new.com"}
	updated, err := client.UpdateAPI(expected.Name, expected)
	assert.Nil(t, err)
	assertAPIEqual(t, expected, updated)

	err = client.DeleteAPI(expected)
	assert.Nil(t, err)
}

func TestDeleteAPI(t *testing.T) {
	dev.EnsureLocal(t)

	client := getTestKongInstance(t)

	noSuchAPI := &gateway.API{
		Name: "testNoSuchAPI",
	}
	err := client.DeleteAPI(noSuchAPI)
	assert.NotNil(t, err)
}
