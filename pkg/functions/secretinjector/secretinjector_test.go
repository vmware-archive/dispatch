///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package secretinjector

import (
	"testing"

	"github.com/go-openapi/strfmt"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/functions/mocks"

	"github.com/vmware/dispatch/pkg/functions"
	secretclient "github.com/vmware/dispatch/pkg/secret-store/gen/client"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
	serviceclient "github.com/vmware/dispatch/pkg/service-manager/gen/client"
	service "github.com/vmware/dispatch/pkg/service-manager/gen/client/service_instance"
	servicemodels "github.com/vmware/dispatch/pkg/service-manager/gen/models"
)

//go:generate mockery -name SecretInjector -case underscore -dir .

func TestInjectSecret(t *testing.T) {

	expectedSecretName := "testSecret"
	expectedSecretValue := models.SecretValue{"secret1": "value1", "secret2": "value2"}

	// expectedSecrets := map[string]interface{}{
	// 	expectedSecretName: expectedSecretValue,
	// }
	expectedOutput := map[string]interface{}{"secret1": "value1", "secret2": "value2"}

	secretTransport := &mocks.ClientTransport{}
	secretTransport.On("Submit", mock.Anything).Return(
		&secret.GetSecretOK{
			Payload: &models.Secret{
				Name:    &expectedSecretName,
				Secrets: expectedSecretValue,
			}}, nil)

	serviceTransport := &mocks.ClientTransport{}
	serviceTransport.On("Submit", mock.Anything).Return(&service.GetServiceInstanceByNameOK{}, nil)

	secretStore := secretclient.New(secretTransport, strfmt.Default)
	serviceManager := serviceclient.New(serviceTransport, strfmt.Default)

	injector := New(secretStore, serviceManager)

	cookie := "testCookie"

	printSecretsFn := func(ctx functions.Context, _ interface{}) (interface{}, error) {
		return ctx["secrets"], nil
	}

	ctx := functions.Context{}
	output, err := injector.GetMiddleware([]string{expectedSecretName}, []string{}, cookie)(printSecretsFn)(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}

func TestInjectService(t *testing.T) {

	expectedSecretValue := models.SecretValue{"secret1": "value1", "secret2": "value2"}
	expectedServiceName := "testService"
	expectedOutput := map[string]interface{}{"secret1": "value1", "secret2": "value2"}

	serviceID := uuid.NewV4().String()

	serviceTransport := &mocks.ClientTransport{}
	serviceTransport.On("Submit", mock.Anything).Return(
		&service.GetServiceInstanceByNameOK{
			Payload: &servicemodels.ServiceInstance{
				Name: &expectedServiceName,
				ID:   strfmt.UUID(serviceID),
				Binding: &servicemodels.ServiceBinding{
					Status: servicemodels.StatusREADY,
				},
			}}, nil)

	secretTransport := &mocks.ClientTransport{}
	secretTransport.On("Submit", mock.Anything).Return(
		&secret.GetSecretOK{
			Payload: &models.Secret{
				Name:    &serviceID,
				Secrets: expectedSecretValue,
			}}, nil)

	secretStore := secretclient.New(secretTransport, strfmt.Default)
	serviceManager := serviceclient.New(serviceTransport, strfmt.Default)

	injector := New(secretStore, serviceManager)

	cookie := "testCookie"

	printServiceFn := func(ctx functions.Context, _ interface{}) (interface{}, error) {
		return ctx["serviceBindings"].(map[string]interface{})[expectedServiceName], nil
	}

	ctx := functions.Context{}
	output, err := injector.GetMiddleware([]string{}, []string{expectedServiceName}, cookie)(printServiceFn)(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}
