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

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/service-manager/entities"
	"github.com/vmware/dispatch/pkg/service-manager/flags"
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
			Status:         entitystore.StatusINITIALIZED,
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
	instance := entities.ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			Name:           "instanceA",
			OrganizationID: "dispatch",
			Status:         entitystore.StatusINITIALIZED,
			Tags: map[string]string{
				"role": "test",
			},
		},
		ServiceClass: "classA",
	}
	binding := entities.ServiceBinding{
		BaseEntity: entitystore.BaseEntity{
			Name:           "instanceA",
			OrganizationID: "dispatch",
			Status:         entitystore.StatusINITIALIZED,
			Tags: map[string]string{
				"role": "test",
			},
		},
		ServiceInstance: "instanceA",
		Parameters: map[string]interface{}{
			"key": "value",
		},
	}

	id, err := handlers.Store.Add(&bindableClass)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	id, err = handlers.Store.Add(&instance)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	id, err = handlers.Store.Add(&binding)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	return map[string]interface{}{
		freePlan.Name:      &freePlan,
		bindableClass.Name: &bindableClass,
		instance.Name:      &instance,
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
	var respBody []v1.ServiceClass
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

	addRequest := v1.ServiceInstance{
		Name:         swag.String("instanceB"),
		ServiceClass: swag.String("classA"),
		Tags: []*v1.Tag{
			&v1.Tag{
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
	var respBody v1.ServiceInstance
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
	assert.NotNil(t, respBody.Binding)
}

func TestGetServiceInstanceByName(t *testing.T) {
	flags.ServiceManagerFlags.OrgID = "dispatch"
	handlers := &Handlers{
		Store: helpers.MakeEntityStore(t),
	}

	api := operations.NewServiceManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	serviceEntities := createServiceEntities(t, handlers)

	r := httptest.NewRequest("GET", "/v1/serviceinstance/instanceA", nil)
	get := serviceinstance.GetServiceInstanceByNameParams{
		HTTPRequest:         r,
		ServiceInstanceName: "instanceA",
	}

	var respBody v1.ServiceInstance
	responder := api.ServiceInstanceGetServiceInstanceByNameHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, responder, &respBody, 200)

	instance := serviceEntities["instanceA"].(*entities.ServiceInstance)

	assert.Equal(t, instance.Name, *respBody.Name)
	assert.Equal(t, string(entitystore.StatusINITIALIZED), string(respBody.Status))
	assert.Equal(t, string(entitystore.StatusINITIALIZED), string(respBody.Binding.Status))
	assert.Equal(t, map[string]interface{}{"key": "value"}, respBody.Binding.Parameters)

	r = httptest.NewRequest("GET", "/v1/serviceinstance/instanceB", nil)
	get = serviceinstance.GetServiceInstanceByNameParams{
		HTTPRequest:         r,
		ServiceInstanceName: "instanceB",
	}
	responder = api.ServiceInstanceGetServiceInstanceByNameHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, responder, &respBody, 404)
}

func TestGetServiceInstances(t *testing.T) {
	flags.ServiceManagerFlags.OrgID = "dispatch"
	handlers := &Handlers{
		Store: helpers.MakeEntityStore(t),
	}

	api := operations.NewServiceManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	r := httptest.NewRequest("GET", "/v1/serviceinstance", nil)
	get := serviceinstance.GetServiceInstancesParams{
		HTTPRequest: r,
	}

	var respBody []*v1.ServiceInstance
	responder := api.ServiceInstanceGetServiceInstancesHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, responder, &respBody, 200)

	assert.Len(t, respBody, 0)

	serviceEntities := createServiceEntities(t, handlers)
	instance := serviceEntities["instanceA"].(*entities.ServiceInstance)

	responder = api.ServiceInstanceGetServiceInstancesHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, responder, &respBody, 200)

	assert.Len(t, respBody, 1)
	assert.Equal(t, instance.Name, *respBody[0].Name)
}

func TestDeleteServiceInstanceByName(t *testing.T) {
	flags.ServiceManagerFlags.OrgID = "dispatch"
	handlers := &Handlers{
		Store: helpers.MakeEntityStore(t),
	}

	api := operations.NewServiceManagerAPI(nil)
	handlers.ConfigureHandlers(api)

	r := httptest.NewRequest("DELETE", "/v1/serviceinstance/instanceA", nil)
	del := serviceinstance.DeleteServiceInstanceByNameParams{
		HTTPRequest:         r,
		ServiceInstanceName: "instanceA",
	}

	var respBody v1.ServiceInstance
	responder := api.ServiceInstanceDeleteServiceInstanceByNameHandler.Handle(del, "testCookie")
	helpers.HandlerRequest(t, responder, &respBody, 404)

	serviceEntities := createServiceEntities(t, handlers)
	instance := serviceEntities["instanceA"].(*entities.ServiceInstance)

	responder = api.ServiceInstanceDeleteServiceInstanceByNameHandler.Handle(del, "testCookie")
	helpers.HandlerRequest(t, responder, &respBody, 200)

	assert.Equal(t, instance.Name, *respBody.Name)

	r = httptest.NewRequest("GET", "/v1/serviceinstance", nil)
	get := serviceinstance.GetServiceInstancesParams{
		HTTPRequest: r,
	}
	var instances []*v1.ServiceInstance
	responder = api.ServiceInstanceGetServiceInstancesHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, responder, &instances, 200)

	assert.Len(t, instances, 0)
}
