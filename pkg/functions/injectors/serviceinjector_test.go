///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package injectors

import (
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client/mocks"
	"github.com/vmware/dispatch/pkg/functions"
)

func TestInjectService(t *testing.T) {

	expectedSecretValue := v1.SecretValue{"secret1": "value1", "secret2": "value2"}
	expectedServiceName := "testService"
	expectedOutput := map[string]interface{}{"secret1": "value1", "secret2": "value2"}

	serviceID := uuid.NewV4().String()

	servicesClient := &mocks.ServicesClient{}
	servicesClient.On("GetServiceInstance", mock.Anything, "testOrg", mock.Anything).Return(
		&v1.ServiceInstance{
			Name: &expectedServiceName,
			ID:   strfmt.UUID(serviceID),
			Binding: &v1.ServiceBinding{
				Status: v1.StatusREADY,
			}}, nil)

	secretsClient := &mocks.SecretsClient{}
	secretsClient.On("GetSecret", mock.Anything, "testOrg", mock.Anything).Return(
		&v1.Secret{
			Name:    &serviceID,
			Secrets: expectedSecretValue,
		}, nil)

	injector := NewServiceInjector(secretsClient, servicesClient)

	cookie := "testCookie"

	printServiceFn := func(ctx functions.Context, _ interface{}) (interface{}, error) {
		return ctx["serviceBindings"].(map[string]interface{})[expectedServiceName], nil
	}

	ctx := functions.Context{}
	output, err := injector.GetMiddleware("testOrg", []string{expectedServiceName}, cookie)(printServiceFn)(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}
