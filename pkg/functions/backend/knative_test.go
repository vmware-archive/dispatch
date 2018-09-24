///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"
	"testing"
	"time"

	"github.com/knative/serving/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/dispatch/pkg/api/v1"
)

const (
	testOrg     = "vmware"
	testProject = "dispatch"

	fn1 = "f1"
	fn2 = "f2"

	secret1 = "secret1"
	secret2 = "secret2"
	secret3 = "secret3"
)

func f1() *v1.Function {
	return &v1.Function{
		Meta: v1.Meta{
			Kind:    v1.FunctionKind,
			Org:     testOrg,
			Project: testProject,
			Name:    fn1,
		},
		FunctionImageURL: "dispatchframework/test-fn1",
		Secrets:          []string{secret1, secret2},
		Image:            "image1",
	}
}

func f2() *v1.Function {
	return &v1.Function{
		Meta: v1.Meta{
			Kind:    v1.FunctionKind,
			Org:     testOrg,
			Project: testProject,
			Name:    fn2,
		},
		FunctionImageURL: "dispatchframework/test-fn2",
		Timeout:          int64(5 * time.Minute),
		Secrets:          []string{secret2, secret3},
		Image:            "image2",
	}
}

func testBackend() *knative {
	buildConfig := &BuildConfig{
		BuildImage:     "dispatchframework/dispatch-knative-builder:0.0.1",
		BuildCommand:   "/fetch_source.sh",
		BuildTemplate:  "function-template",
		ServiceAccount: "dispatch-build",
	}
	be := &knative{
		knClient:    fake.NewSimpleClientset(),
		buildConfig: buildConfig,
	}
	return be
}

func setup(t *testing.T) Backend {
	be := testBackend()

	var err error
	_, err = be.Add(context.TODO(), f1())
	require.NoError(t, err)
	_, err = be.Add(context.TODO(), f2())
	require.NoError(t, err)
	return be
}

func TestKnative_AddGet(t *testing.T) {
	be := setup(t)

	fun1, err := be.Get(context.TODO(), &v1.Meta{Org: testOrg, Project: testProject, Name: fn1})
	require.NoError(t, err)
	require.NotEmpty(t, fun1)

	assert.NotEmpty(t, fun1.Meta.BackingObject)
	assert.NotEmpty(t, fun1.Meta.CreatedTime)

	fun1.Meta.CreatedTime = 0
	fun1.Meta.BackingObject = nil
	fun1.ModifiedTime = 0
	fun1.Status = ""

	assert.Equal(t, f1(), fun1)
}

func TestKnative_List(t *testing.T) {
	be := setup(t)

	functions, err := be.List(context.TODO(), &v1.Meta{Org: testOrg, Project: testProject})
	require.NoError(t, err)
	require.NotEmpty(t, functions)

	assert.Equal(t, 2, len(functions))
}

func TestKnative_Update(t *testing.T) {
	be := setup(t)

	const newImage = "dispatchframework/test-fn2:v0.2"

	f2NewImage := f2()
	f2NewImage.FunctionImageURL = newImage

	function, err := be.Update(context.TODO(), f2NewImage)
	require.NoError(t, err)
	require.NotEmpty(t, function)

	assert.Equal(t, newImage, function.FunctionImageURL)

	function, err = be.Get(context.TODO(), &f2NewImage.Meta)
	require.NoError(t, err)
	require.NotEmpty(t, function)

	assert.Equal(t, newImage, function.FunctionImageURL)
}

func TestKnative_Delete(t *testing.T) {
	be := setup(t)

	err := be.Delete(context.TODO(), &v1.Meta{Org: testOrg, Project: testProject, Name: fn1})
	require.NoError(t, err)

	functions, err := be.List(context.TODO(), &v1.Meta{Org: testOrg, Project: testProject})
	require.NoError(t, err)

	assert.Equal(t, 1, len(functions))
	assert.Equal(t, fn2, functions[0].Meta.Name)
}
