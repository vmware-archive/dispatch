///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package openfaas

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sFake "k8s.io/client-go/kubernetes/fake"

	"github.com/openfaas/faas/gateway/requests"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/testing/dev"
	"k8s.io/api/apps/v1beta1"
)

func driver() *ofDriver {
	log.SetLevel(log.DebugLevel)

	d, err := New(&Config{
		Gateway: "http://localhost:8080/",
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "driver instance not created"))
	}
	return d.(*ofDriver)
}

func TestOfDriverCreate(t *testing.T) {
	testHttpserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		var req requests.CreateFunctionRequest
		json.Unmarshal(reqBody, &req)
		assert.Equal(t, "fake-image:latest", req.Image)
		assert.Nil(t, req.Requests)
		assert.Nil(t, req.Limits)
		w.WriteHeader(http.StatusOK)
	}))

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
		FunctionImageURL: "fake-image:latest",
	}

	deploymentObj := &v1beta1.Deployment{
		TypeMeta: k8sMetaV1.TypeMeta{},
		ObjectMeta: k8sMetaV1.ObjectMeta{
			Namespace: "fakeNS",
			Name:      getID(f.FaasID),
		},
		Spec: v1beta1.DeploymentSpec{},
		Status: v1beta1.DeploymentStatus{
			AvailableReplicas: 1,
		},
	}

	clientSet := k8sFake.NewSimpleClientset(deploymentObj)

	fakeAppsV1beta1 := clientSet.AppsV1beta1()
	fakeDeployments := fakeAppsV1beta1.Deployments("fakeNS")

	d := ofDriver{
		gateway:       testHttpserver.URL,
		httpClient:    testHttpserver.Client(),
		createTimeout: defaultCreateTimeout,
		deployments:   fakeDeployments,
	}

	err := d.Create(context.Background(), &f)
	assert.NoError(t, err)
}

func TestOfDriver_GetRunnable(t *testing.T) {
	dev.EnsureLocal(t)

	d := driver()

	f := d.GetRunnable(&functions.FunctionExecution{FunctionID: "deadbeef"})
	ctx := functions.Context{}
	r, err := f(ctx, map[string]interface{}{"name": "Me", "place": "Here"})

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"myField": "Hello, Me from Here"}, r)
	assert.Equal(t, []string{"log log log", "log log log"}, ctx["logs"])
}

func TestDriver_Create(t *testing.T) {
	dev.EnsureLocal(t)

	d := driver()

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
		FunctionImageURL: "testfunc",
	}

	err := d.Create(context.Background(), &f)
	assert.NoError(t, err)
}

func TestOfDriver_Delete(t *testing.T) {
	dev.EnsureLocal(t)

	d := driver()

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}
	err := d.Delete(context.Background(), &f)
	assert.NoError(t, err)
}
