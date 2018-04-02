///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package servicemanager

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/service-manager/entities"
	"github.com/vmware/dispatch/pkg/service-manager/flags"
	"github.com/vmware/dispatch/pkg/service-manager/gen/models"
	"github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations"
	serviceclass "github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations/service_class"
	serviceinstance "github.com/vmware/dispatch/pkg/service-manager/gen/restapi/operations/service_instance"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func TestGetServiceClassByName(t *testing.T) {

}

func createServiceEntities(t *testing.T, handlers *Handlers) map[string]interface{} {
	freePlan := entities.ServicePlan{
		BaseEntity: entitystore.BaseEntity{
			Name:           "planA",
			OrganizationID: "dispatch",
		},
		Description: "planA is for testing",
		Schema: entities.Schema{
			Create: &spec.Schema{},
			Update: &spec.Schema{},
			Bind:   &spec.Schema{},
		},
		Free: true,
	}
	bindableClass := entities.ServiceClass{
		BaseEntity: entitystore.BaseEntity{
			Name:           "classA",
			OrganizationID: "dispatch",
			Tags: map[string]string{
				"role": "test",
			},
		},
		Description: "classA is for testing",
		ServiceID:   "deadbeef",
		Broker:      "brokerA",
		Bindable:    true,
		Plans:       []entities.ServicePlan{freePlan},
	}

	id, err := handlers.Store.Add(&bindableClass)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	return map[string]interface{}{
		freePlan.Name:      &freePlan,
		bindableClass.Name: &bindableClass,
	}
}

func TestGetServiceClasses(t *testing.T) {
	flags.ServiceManagerFlags.OrgID = "dispatch"
	handlers := &Handlers{
		Store: helpers.MakeEntityStore(t),
	}

	api := operations.NewServiceManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	serviceEntities := createServiceEntities(t, handlers)
	bindableClass := serviceEntities["classA"].(*entities.ServiceClass)
	freePlan := serviceEntities["planA"].(*entities.ServicePlan)

	r := httptest.NewRequest("GET", "/v1/serviceclass", nil)
	get := serviceclass.GetServiceClassesParams{
		HTTPRequest: r,
	}
	responder := api.ServiceClassGetServiceClassesHandler.Handle(get, "testCookie")
	var respBody []models.ServiceClass
	helpers.HandlerRequest(t, responder, &respBody, 200)

	t.Logf("response: %+v", respBody)

	assert.Len(t, respBody, 1)
	respBindable := respBody[0]
	assert.Equal(t, bindableClass.Name, *respBindable.Name)
	assert.Equal(t, bindableClass.Broker, *respBindable.Broker)
	assert.Equal(t, bindableClass.Bindable, respBindable.Bindable)
	assert.Equal(t, "role", respBindable.Tags[0].Key)
	assert.Equal(t, "test", respBindable.Tags[0].Value)
	assert.Len(t, respBindable.Plans, 1)
	respPlan := respBindable.Plans[0]
	assert.Equal(t, freePlan.Name, respPlan.Name)
	assert.Equal(t, freePlan.Free, respPlan.Free)
	assert.Equal(t, freePlan.Bindable, respPlan.Bindable)
}

func TestAddServiceInstance(t *testing.T) {
	flags.ServiceManagerFlags.OrgID = "dispatch"
	handlers := &Handlers{
		Store: helpers.MakeEntityStore(t),
	}

	api := operations.NewServiceManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	addRequest := models.ServiceInstance{
		Name:         swag.String("instanceA"),
		ServiceClass: swag.String("classA"),
		Tags: []*models.Tag{
			&models.Tag{
				Key:   "role",
				Value: "test",
			},
		},
		SecretParameters: []string{"secretA"},
		Parameters: map[string]string{
			"keyA": "valueA",
		},
		ServicePlan: swag.String("planA"),
	}
	r := httptest.NewRequest("POST", "/v1/serviceinstance", nil)
	post := serviceinstance.AddServiceInstanceParams{
		HTTPRequest: r,
		Body:        &addRequest,
	}
	responder := api.ServiceInstanceAddServiceInstanceHandler.Handle(post, "testCookie")
	var respBody models.ServiceInstance
	// No service class defined
	helpers.HandlerRequest(t, responder, &respBody, 400)

	_ = createServiceEntities(t, handlers)
	responder = api.ServiceInstanceAddServiceInstanceHandler.Handle(post, "testCookie")
	// Service class and plan now exist
	helpers.HandlerRequest(t, responder, &respBody, 201)

	assert.Equal(t, addRequest.Name, respBody.Name)
	assert.Equal(t, addRequest.ServicePlan, respBody.ServicePlan)
	assert.Len(t, respBody.SecretParameters, 1)
	assert.Equal(t, addRequest.SecretParameters, respBody.SecretParameters)
}

func TestGetServiceInstanceByName(t *testing.T) {

}

func TestGetServiceInstances(t *testing.T) {

}

func TestDeleteServiceInstanceByName(t *testing.T) {

}
