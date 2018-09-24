///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"
	"testing"

	"github.com/knative/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/dispatch/pkg/api/v1"
)

const (
	testOrg     = "vmware"
	testProject = "dispatch"

	en1 = "e1"
	en2 = "e2"
)

func e1() *v1.Endpoint {
	return &v1.Endpoint{
		Meta: v1.Meta{
			Kind:    v1.EndpointKind,
			Org:     testOrg,
			Project: testProject,
			Name:    en1,
		},
		Function: "test-fn1",
		Uris:     []string{"/test-fn1"},
		Methods:  []string{"GET", "POST"},
		Hosts:    []string{"dispatch.vmware.test.dispatch.local"},
	}
}

func e2() *v1.Endpoint {
	return &v1.Endpoint{
		Meta: v1.Meta{
			Kind:    v1.EndpointKind,
			Org:     testOrg,
			Project: testProject,
			Name:    en2,
		},
		Function: "test-fn2",
		Uris:     []string{"/test-fn2"},
		Methods:  []string{"GET"},
	}
}

func testBackend() *knative {
	be := &knative{
		knClient: fake.NewSimpleClientset(),
		config: knativeEndpointsConfig{
			InternalGateway: "knative-ingressgateway.istio-system.svc.cluster.local",
			SharedGateway:   "knative-shared-gateway.knative-serving.svc.cluster.local",
			DispatchHost:    "test.dispatch.local",
		},
	}
	return be
}

func setup(t *testing.T) Backend {
	be := testBackend()

	var err error
	_, err = be.Add(context.TODO(), e1())
	require.NoError(t, err)
	_, err = be.Add(context.TODO(), e2())
	require.NoError(t, err)
	return be
}

func TestKnative_AddGet(t *testing.T) {
	be := setup(t)

	end1, err := be.Get(context.TODO(), &v1.Meta{Org: testOrg, Project: testProject, Name: en1})
	require.NoError(t, err)
	require.NotEmpty(t, end1)

	assert.NotEmpty(t, end1.Meta.BackingObject)
	assert.NotEmpty(t, end1.Meta.CreatedTime)

	end1.Meta.CreatedTime = 0
	end1.Meta.BackingObject = nil
	end1.ModifiedTime = 0
	end1.Status = ""

	assert.Equal(t, e1(), end1)
}

func TestKnative_List(t *testing.T) {
	be := setup(t)

	endpoints, err := be.List(context.TODO(), &v1.Meta{Org: testOrg, Project: testProject})
	require.NoError(t, err)
	require.NotEmpty(t, endpoints)

	assert.Equal(t, 2, len(endpoints))
}

func TestKnative_Update(t *testing.T) {
	be := setup(t)

	const newFunction = "new-fn2"

	e2NewFunction := e2()
	e2NewFunction.Function = newFunction

	updated, err := be.Update(context.TODO(), e2NewFunction)
	require.NoError(t, err)
	require.NotEmpty(t, updated)

	assert.Equal(t, newFunction, updated.Function)

	updated, err = be.Get(context.TODO(), &e2NewFunction.Meta)
	require.NoError(t, err)
	require.NotEmpty(t, updated)

	assert.Equal(t, newFunction, updated.Function)
}

func TestKnative_Delete(t *testing.T) {
	be := setup(t)

	err := be.Delete(context.TODO(), &v1.Meta{Org: testOrg, Project: testProject, Name: en1})
	require.NoError(t, err)

	endpoints, err := be.List(context.TODO(), &v1.Meta{Org: testOrg, Project: testProject})
	require.NoError(t, err)

	assert.Equal(t, 1, len(endpoints))
	assert.Equal(t, en2, endpoints[0].Meta.Name)
}
