///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package injectors

import (
	"testing"

	"github.com/go-openapi/strfmt"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/functions/mocks"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/functions"
	secretclient "github.com/vmware/dispatch/pkg/secret-store/gen/client"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	serviceclient "github.com/vmware/dispatch/pkg/service-manager/gen/client"
	service "github.com/vmware/dispatch/pkg/service-manager/gen/client/service_instance"
)

//go:generate mockery -name ServiceInjector -case underscore -dir .

func TestInjectService(t *testing.T) {

	expectedSecretValue := v1.SecretValue{"secret1": "value1", "secret2": "value2"}
	expectedServiceName := "testService"
	expectedOutput := map[string]interface{}{"secret1": "value1", "secret2": "value2"}

	serviceID := uuid.NewV4().String()

	serviceTransport := &mocks.ClientTransport{}
	serviceTransport.On("Submit", mock.Anything).Return(
		&service.GetServiceInstanceByNameOK{
			Payload: &v1.ServiceInstance{
				Name: &expectedServiceName,
				ID:   strfmt.UUID(serviceID),
				Binding: &v1.ServiceBinding{
					Status: v1.StatusREADY,
				},
			}}, nil)

	secretTransport := &mocks.ClientTransport{}
	secretTransport.On("Submit", mock.Anything).Return(
		&secret.GetSecretOK{
			Payload: &v1.Secret{
				Name:    &serviceID,
				Secrets: expectedSecretValue,
			}}, nil)

	secretStore := secretclient.New(secretTransport, strfmt.Default)
	serviceManager := serviceclient.New(serviceTransport, strfmt.Default)

	injector := NewServiceInjector(secretStore, serviceManager)

	cookie := "testCookie"

	printServiceFn := func(ctx functions.Context, _ interface{}) (interface{}, error) {
		return ctx["serviceBindings"].(map[string]interface{})[expectedServiceName], nil
	}

	ctx := functions.Context{}
	output, err := injector.GetMiddleware([]string{expectedServiceName}, cookie)(printServiceFn)(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}
