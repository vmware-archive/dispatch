///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package openfaas

import (
	"context"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/testing/dev"
)

func registryAuth() string {
	return os.Getenv("REGISTRY_AUTH")
}

func driver() *ofDriver {
	log.SetLevel(log.DebugLevel)

	d, err := New(&Config{
		Gateway:       "http://localhost:8080/",
		ImageRegistry: "serverless-docker-local.artifactory.eng.vmware.com",
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

	f := d.GetRunnable("hello")
	r, err := f(map[string]interface{}{"name": "Me", "place": "Here"})

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"myField": "Hello, Me from Here"}, r)
}

func TestDriver_Create(t *testing.T) {
	dev.EnsureLocal(t)

	require.NotEmpty(t, registryAuth())
	d := driver()
	defer d.Shutdown()

	err := d.Create("hello", &functions.Exec{
		Image: "serverless-docker-local.artifactory.eng.vmware.com/photon-func-deps-node:7.7.4",
		Code: `
module.exports = function (params) {
    let name = "Noone";
    if (params.name) {
        name = params.name;
    }
    let place = "Nowhere";
    if (params.place) {
        place = params.place;
    }
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

	err := d.Delete("hello")
	assert.NoError(t, err)
}
