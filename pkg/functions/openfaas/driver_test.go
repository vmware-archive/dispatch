///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package openfaas

import (
	"context"
	"os"
	"testing"

	"github.com/vmware/dispatch/pkg/entity-store"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/testing/dev"
)

func registryAuth() string {
	return os.Getenv("REGISTRY_AUTH")
}

func driver() *ofDriver {
	log.SetLevel(log.DebugLevel)

	d, err := New(&Config{
		Gateway:       "http://localhost:8080/",
		ImageRegistry: "vmware",
		RegistryAuth:  registryAuth(),
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "driver instance not created"))
	}
	return d.(*ofDriver)
}

func TestImagePull(t *testing.T) {
	dev.EnsureLocal(t)

	require.NotEmpty(t, registryAuth())

	d := driver()
	defer d.Shutdown()

	err := dockerError(
		d.docker.ImagePull(context.Background(), "imikushin/no-such-mf-image", types.ImagePullOptions{}),
	)
	assert.Error(t, err)
}

func TestImagePush(t *testing.T) {
	dev.EnsureLocal(t)

	require.NotEmpty(t, registryAuth())

	d := driver()
	defer d.Shutdown()

	err := dockerError(
		d.docker.ImagePush(context.Background(), "imikushin/no-such-mf-image", types.ImagePushOptions{
			RegistryAuth: registryAuth(),
		}),
	)
	assert.Error(t, err)
}

func TestOfDriver_GetRunnable(t *testing.T) {
	dev.EnsureLocal(t)

	d := driver()
	defer d.Shutdown()

	f := d.GetRunnable(&functions.FunctionExecution{Name: "hello", ID: "deadbeef"})
	ctx := functions.Context{}
	r, err := f(ctx, map[string]interface{}{"name": "Me", "place": "Here"})

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"myField": "Hello, Me from Here"}, r)
	assert.Equal(t, []string{"log log log", "log log log"}, ctx["logs"])
}

func TestDriver_Create(t *testing.T) {
	dev.EnsureLocal(t)

	require.NotEmpty(t, registryAuth())
	d := driver()
	defer d.Shutdown()

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}

	err := d.Create(&f, &functions.Exec{
		Image:    "vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1",
		Language: "nodejs6",
		Code: `
module.exports = function (context, input) {
    let name = "Noone";
    if (input.name) {
        name = input.name;
    }
    let place = "Nowhere";
    if (input.place) {
        place = input.place;
    }
    console.log('log log log');
    console.log('log log log');
    return {myField:  'Hello, ' + name + ' from ' + place}
};
`,
	})
	assert.NoError(t, err)
}

func TestOfDriver_Delete(t *testing.T) {
	dev.EnsureLocal(t)

	d := driver()
	defer d.Shutdown()

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}
	err := d.Delete(&f)
	assert.NoError(t, err)
}
