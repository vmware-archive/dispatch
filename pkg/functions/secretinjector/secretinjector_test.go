///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package secretinjector

import (
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/functions/mocks"

	"github.com/vmware/dispatch/pkg/functions"
	secretclient "github.com/vmware/dispatch/pkg/secret-store/gen/client"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	"github.com/vmware/dispatch/pkg/secret-store/gen/models"
)

func TestImpl_GetMiddleware(t *testing.T) {

	expectedSecretName := "testSecret"
	expectedSecretValue := models.SecretValue{"secret1": "value1", "secret2": "value2"}
	expectedSecrets := map[string]interface{}{
		expectedSecretName: expectedSecretValue,
	}
	expectedOutput := map[string]interface{}{"secret1": "value1", "secret2": "value2"}

	mockedTransport := &mocks.ClientTransport{}
	mockedTransport.On("Submit", mock.Anything).Return(
		&secret.GetSecretOK{
			Payload: &models.Secret{
				Name:    &expectedSecretName,
				Secrets: expectedSecretValue,
			}}, nil)

	secretStore := secretclient.New(mockedTransport, strfmt.Default)

	injector := New(secretStore)

	cookie := "testCookie"

	printSecretsFn := func(ctx functions.Context, _ interface{}) (interface{}, error) {
		return ctx["secrets"], nil
	}

	ctx := functions.Context{"secrets": expectedSecrets}
	output, err := injector.GetMiddleware([]string{expectedSecretName}, cookie)(printSecretsFn)(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}
