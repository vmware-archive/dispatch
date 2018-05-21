///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package injectors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/client/mocks"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/functions"
)

func TestInjectSecret(t *testing.T) {

	expectedSecretName := "testSecret"
	expectedSecretValue := v1.SecretValue{"secret1": "value1", "secret2": "value2"}

	expectedOutput := map[string]interface{}{"secret1": "value1", "secret2": "value2"}

	secretsClient := &mocks.SecretsClient{}
	secretsClient.On("GetSecret", mock.Anything, "testOrg", mock.Anything).Return(
		&v1.Secret{
			Name:    &expectedSecretName,
			Secrets: expectedSecretValue,
		}, nil)

	injector := NewSecretInjector(secretsClient)

	cookie := "testCookie"

	printSecretsFn := func(ctx functions.Context, _ interface{}) (interface{}, error) {
		return ctx["secrets"], nil
	}

	ctx := functions.Context{}
	output, err := injector.GetMiddleware("testOrg", []string{expectedSecretName}, cookie)(printSecretsFn)(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}
