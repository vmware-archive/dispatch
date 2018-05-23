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

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/testing/dev"
)

const (
	funID   = "cafebabe-face-abba"
	funName = "hello"
)

func driver() *riffDriver {
	log.SetLevel(log.DebugLevel)

	d, err := New(&Config{
		K8sConfig:     os.Getenv("K8S_CONFIG"),
		FuncNamespace: "riff",
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "driver instance not created"))
	}
	return d.(*riffDriver)
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

	d := driver()

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: funName,
			ID:   funID,
		},
		FunctionImageURL: "testfunc",
	}

	err := d.Create(context.Background(), &f)
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
	err := d.Delete(context.Background(), &f)
	assert.NoError(t, err)
}
