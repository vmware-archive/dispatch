///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package openwhisk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/testing/dev"
)

var driver functions.FaaSDriver

func init() {
	d, err := New(&Config{
		Insecure:  true,
		Host:      "52.91.175.16",
		AuthToken: "23bc46b1-71f6-4ed5-8c54-816aa4f8c502:123zO3xZCLrMN6v2BKK1dXYFpXlPkccOFqm12CdAsMgRU4VrNZ9lyGVCGuMDGIwP",
	})
	if err != nil {
		panic(err)
	}
	driver = d
}

func TestWskDriver_GetRunnable(t *testing.T) {
	dev.EnsureLocal(t)
	f := driver.GetRunnable(&functions.FunctionExecution{Name: "hello", ID: "deadbeef"})
	r, err := f(functions.Context{}, map[string]interface{}{})

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"myField": "Hello, Noone from Nowhere"}, r)
}

func TestWskDriver_Delete(t *testing.T) {
	dev.EnsureLocal(t)
	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}
	err := driver.Delete(&f)
	assert.NoError(t, err)
}

func TestWskDriver_Create(t *testing.T) {
	dev.EnsureLocal(t)
	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}
	err := driver.Create(&f, &functions.Exec{
		Image: "imikushin/nodejs6action",
		Code: `
function main(ctx, params) {
    let name = "Noone";
    if (params.name) {
        name = params.name;
    }
    let place = "Nowhere";
    if (params.place) {
        place = params.place;
    }
    return {myField:  'Hello, ' + name + ' from ' + place};
}
		`,
	})
	assert.NoError(t, err)
}
