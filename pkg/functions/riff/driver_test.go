///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package riff

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
	"github.com/vmware/dispatch/pkg/images"
	"github.com/vmware/dispatch/pkg/testing/dev"
)

const (
	funID   = "cafebabe-face-abba"
	funName = "hello"
)

func registryAuth() string {
	return os.Getenv("REGISTRY_AUTH")
}

func driver() *riffDriver {
	log.SetLevel(log.DebugLevel)

	d, err := New(&Config{
		ImageRegistry: "imikushin",
		RegistryAuth:  registryAuth(),
		K8sConfig:     os.Getenv("K8S_CONFIG"),
		FuncNamespace: "riff",
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "driver instance not created"))
	}
	return d.(*riffDriver)
}

func TestImagePull(t *testing.T) {
	dev.EnsureLocal(t)

	require.NotEmpty(t, registryAuth())

	d := driver()

	err := images.DockerError(
		d.docker.ImagePull(context.Background(), "imikushin/no-such-mf-image", types.ImagePullOptions{}),
	)
	assert.Error(t, err)
}

func TestImagePush(t *testing.T) {
	dev.EnsureLocal(t)

	require.NotEmpty(t, registryAuth())

	d := driver()

	err := images.DockerError(
		d.docker.ImagePush(context.Background(), "imikushin/no-such-mf-image", types.ImagePushOptions{
			RegistryAuth: registryAuth(),
		}),
	)
	assert.Error(t, err)
}

func TestDriver_GetRunnable(t *testing.T) {
	dev.EnsureLocal(t)

	d := driver()

	f := d.GetRunnable(&functions.FunctionExecution{FunctionID: funID})
	ctx := functions.Context{}
	r, err := f(ctx, map[string]interface{}{"name": "Noone", "place": "Braavos"})

	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"greeting": "Hello, Noone from Braavos"}, r)
	assert.Equal(t, []string{"log log log", "log log log"}, ctx.Logs())
}

func TestDriver_Create(t *testing.T) {
	dev.EnsureLocal(t)

	require.NotEmpty(t, registryAuth())
	d := driver()

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: funName,
			ID:   funID,
		},
	}

	err := d.Create(&f, &functions.Exec{
		Image: "imikushin/dispatch-riff-nodejs6-base:0.0.3-dev1",
		Code: `
module.exports = (context, {name, place}) => {
    if (!name) {
        name = "Someone";
    }
    if (!place) {
        place = "Somewhere";
    }
    console.log('log log log');
    console.log('log log log');
    return {greeting: 'Hello, ' + name + ' from ' + place}
};
`,
	})
	assert.NoError(t, err)
}

func TestDriver_Delete(t *testing.T) {
	dev.EnsureLocal(t)

	d := driver()

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: funName,
			ID:   funID,
		},
	}
	err := d.Delete(&f)
	assert.NoError(t, err)
}
