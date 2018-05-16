///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package kubeless

import (
	"context"
	"testing"

	kubelessFake "github.com/kubeless/kubeless/pkg/client/clientset/versioned/fake"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sFake "k8s.io/client-go/kubernetes/fake"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/functions"
	"github.com/vmware/dispatch/pkg/testing/dev"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

func driver() *kubelessDriver {
	log.SetLevel(log.DebugLevel)

	d, err := New(&Config{})
	if err != nil {
		log.Fatal(errors.Wrap(err, "driver instance not created"))
	}
	return d.(*kubelessDriver)
}

func TestOfDriverCreate(t *testing.T) {
	f := functions.Function{
		BaseEntity: entitystore.BaseEntity{
			Name: "hello",
			ID:   "deadbeef",
		},
	}

	deploymentObj := &extensionsv1beta1.Deployment{
		TypeMeta: k8sMetaV1.TypeMeta{},
		ObjectMeta: k8sMetaV1.ObjectMeta{
			Namespace: "fakeNS",
			Name:      getID(f.FaasID),
		},
		Spec: extensionsv1beta1.DeploymentSpec{},
		Status: extensionsv1beta1.DeploymentStatus{
			AvailableReplicas: 1,
		},
	}

	clientSet := k8sFake.NewSimpleClientset(deploymentObj)
	kubeCli := kubelessFake.NewSimpleClientset()

	d := kubelessDriver{
		createTimeout: defaultCreateTimeout,
		deployments:   clientSet.ExtensionsV1beta1().Deployments("fakeNS"),
		functions:     kubeCli.KubelessV1beta1().Functions("fakeNS"),
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
