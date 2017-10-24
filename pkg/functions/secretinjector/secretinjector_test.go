///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package secretinjector

import (
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions/mocks"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
	secretclient "gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/client"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/client/secret"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/models"
)

func TestImpl_GetMiddleware(t *testing.T) {

	expectedSecretName := "testSecret"
	expectedSecretValue := models.SecretValue{"secret1": "value1", "secret2": "value2"}
	expectedSecrets := map[string]interface{}{
		expectedSecretName: expectedSecretValue,
	}

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
	assert.Equal(t, expectedSecrets, output)
}
