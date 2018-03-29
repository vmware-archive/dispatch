///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package openfaas

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sFake "k8s.io/client-go/kubernetes/fake"

	"encoding/json"
	"github.com/openfaas/faas/gateway/requests"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/functions/mocks"
	"github.com/vmware/dispatch/pkg/images"
	"github.com/vmware/dispatch/pkg/testing/dev"
	"io/ioutil"
	"k8s.io/api/apps/v1beta1"
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

func TestOfDriverCreate(t *testing.T) {
	mockImageBuilder := &mocks.ImageBuilder{}
	mockImageBuilder.On("BuildImage", mock.Anything, mock.Anything, mock.Anything).Return("fake-image:latest", nil)
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
	}

	deploymentObj := &v1beta1.Deployment{
		TypeMeta: k8sMetaV1.TypeMeta{},
		ObjectMeta: k8sMetaV1.ObjectMeta{
			Namespace: "fakeNS",
			Name:      getID(f.ID),
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
		imageBuilder:  mockImageBuilder,
		httpClient:    testHttpserver.Client(),
		createTimeout: defaultCreateTimeout,
		deployments:   fakeDeployments,
	}

	err := d.Create(&f, &functions.Exec{})
	assert.NoError(t, err)
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

	require.NotEmpty(t, registryAuth())
	d := driver()

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}

	err := d.Create(&f, &functions.Exec{
		Image:    "vmware/dispatch-nodejs6-base:0.0.1-dev1",
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

	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}
	err := d.Delete(&f)
	assert.NoError(t, err)
}
