///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package openfaas

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/testing/dev"
)

var (
	driver functions.FaaSDriver
)

func init() {
	log.SetLevel(log.DebugLevel)

	driver = New(&Config{Gateway: "http://localhost:8080/"})
}

func TestOfDriver_GetRunnable(t *testing.T) {
	dev.EnsureLocal(t)
	f := driver.GetRunnable("hello")
	r, err := f(map[string]interface{}{"name": "Me", "place": "Here"})

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"myField": "Hello, Me from Here"}, r)
}

func TestDriver_Create(t *testing.T) {
	dev.EnsureLocal(t)
	err := driver.Create("hello", &functions.Exec{
		Image: "hello:latest",
	})
	assert.NoError(t, err)
}

func TestOfDriver_Delete(t *testing.T) {
	dev.EnsureLocal(t)
	err := driver.Delete("hello")
	assert.NoError(t, err)
}
