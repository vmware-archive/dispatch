///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	endpoint "github.com/vmware/dispatch/pkg/api-manager/gen/client/endpoint"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	runner "github.com/vmware/dispatch/pkg/function-manager/gen/client/runner"
	function "github.com/vmware/dispatch/pkg/function-manager/gen/client/store"
	policy "github.com/vmware/dispatch/pkg/identity-manager/gen/client/policy"
	serviceaccount "github.com/vmware/dispatch/pkg/identity-manager/gen/client/serviceaccount"
	baseimage "github.com/vmware/dispatch/pkg/image-manager/gen/client/base_image"
	image "github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	secret "github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
)

func msg(m *string) string {
	if m == nil {
		return ""
	}
	return *m
}

func formatCliError(err error, message string) error {
	return i18n.Errorf("CLI Error: %s", message)
}

func formatAPIError(err error, params interface{}) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	// BaseImage
	// Add
	case *baseimage.AddBaseImageBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *baseimage.AddBaseImageConflict:
		return i18n.Errorf("[Code: %d] Conflict: %s", v.Payload.Code, msg(v.Payload.Message))
	case *baseimage.AddBaseImageDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *baseimage.UpdateBaseImageByNameNotFound:
		p := params.(*baseimage.UpdateBaseImageByNameParams)
		return i18n.Errorf("[Code: %d] Base image not found: %s", 404, p.BaseImageName)
	case *baseimage.UpdateBaseImageByNameBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	// Delete
	case *baseimage.DeleteBaseImageByNameBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *baseimage.DeleteBaseImageByNameNotFound:
		p := params.(*baseimage.DeleteBaseImageByNameParams)
		return i18n.Errorf("[Code: %d] Base image not found: %s", v.Payload.Code, p.BaseImageName)
	case *baseimage.DeleteBaseImageByNameDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Get
	case *baseimage.GetBaseImageByNameBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *baseimage.GetBaseImageByNameNotFound:
		p := params.(*baseimage.GetBaseImageByNameParams)
		return i18n.Errorf("[Code: %d] Base image not found: %s", v.Payload.Code, p.BaseImageName)
	case *baseimage.GetBaseImageByNameDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// List
	case *baseimage.GetBaseImagesDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Image
	// Add
	case *image.AddImageBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *image.AddImageConflict:
		return i18n.Errorf("[Code: %d] Conflict: %s", v.Payload.Code, msg(v.Payload.Message))
	case *image.AddImageDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Update
	case *image.UpdateImageByNameBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *image.UpdateImageByNameNotFound:
		p := params.(*image.UpdateImageByNameParams)
		return i18n.Errorf("[Code: %d] Image not found: %s", v.Payload.Code, p.ImageName)
	case *image.UpdateImageByNameDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Delete
	case *image.DeleteImageByNameBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *image.DeleteImageByNameNotFound:
		p := params.(*image.DeleteImageByNameParams)
		return i18n.Errorf("[Code: %d] Image not found: %s", v.Payload.Code, p.ImageName)
	case *image.DeleteImageByNameDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Get
	case *image.GetImageByNameBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *image.GetImageByNameNotFound:
		p := params.(*image.GetImageByNameParams)
		return i18n.Errorf("[Code: %d] Image not found: %s", v.Payload.Code, p.ImageName)
	case *image.GetImageByNameDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// List
	case *image.GetImagesDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Function
	// Add
	case *function.AddFunctionBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *function.AddFunctionConflict:
		return i18n.Errorf("[Code: %d] Conflict: %s", v.Payload.Code, msg(v.Payload.Message))
	case *function.AddFunctionUnauthorized:
		return i18n.Errorf("[Code: %d] Unauthorized: %s", v.Payload.Code, msg(v.Payload.Message))
	case *function.AddFunctionInternalServerError:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Update
	case *function.UpdateFunctionBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *function.UpdateFunctionNotFound:
		p := params.(*function.UpdateFunctionParams)
		return i18n.Errorf("[Code: %d] Function not found: %s", v.Payload.Code, p.FunctionName)
	case *function.UpdateFunctionInternalServerError:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Delete
	case *function.DeleteFunctionBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *function.DeleteFunctionNotFound:
		p := params.(*function.DeleteFunctionParams)
		return i18n.Errorf("[Code: %d] Function not found: %s", v.Payload.Code, p.FunctionName)
	// Get
	case *function.GetFunctionBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *function.GetFunctionNotFound:
		p := params.(*function.GetFunctionParams)
		return i18n.Errorf("[Code: %d] Function not found: %s", v.Payload.Code, p.FunctionName)
	// List
	case *function.GetFunctionsDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Runner
	// Get
	case *runner.GetRunNotFound:
		p := params.(*runner.GetRunParams)
		return i18n.Errorf("[Code: %d] Function execution not found: %s", v.Payload.Code, p.RunName)
	// Exec
	case *runner.RunFunctionBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *runner.RunFunctionNotFound:
		p := params.(*runner.RunFunctionParams)
		return i18n.Errorf("[Code: %d] Function execution not found: %s", v.Payload.Code, *p.FunctionName)
	case *runner.RunFunctionInternalServerError:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *runner.RunFunctionBadGateway:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// List
	case *runner.GetRunsNotFound:
		p := params.(*runner.GetRunsParams)
		return i18n.Errorf("[Code: %d] Function executions not found: %s", v.Payload.Code, *p.FunctionName)
		// Secret
	// Get
	case *secret.GetSecretNotFound:
		return i18n.Errorf("[Code: %d] get Secret not found: %s", v.Payload.Code, msg(v.Payload.Message))
	case *secret.GetSecretDefault:
		return i18n.Errorf("[Code: %d] get Secret error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *secret.GetSecretsDefault:
		return i18n.Errorf("[Code: %d] get Secret error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Update
	case *secret.UpdateSecretDefault:
		return i18n.Errorf("[Code: %d] update Secret error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *secret.UpdateSecretNotFound:
		return i18n.Errorf("[Code: %d] update Secret error: %s", 404, "Secret not found")
	// Create
	case *secret.AddSecretConflict:
		return i18n.Errorf("[Code: %d] Conflict: %s", v.Payload.Code, msg(v.Payload.Message))
	case *secret.AddSecretDefault:
		return i18n.Errorf("[Code: %d] create Secret error: %s", v.Payload.Code, msg(v.Payload.Message))

	// API
	// List
	case *endpoint.GetApisInternalServerError:
		return i18n.Errorf("[Code: %d] get apis error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Get
	case *endpoint.GetAPIBadRequest:
		return i18n.Errorf("[Code: %d] get api error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *endpoint.AddAPIConflict:
		return i18n.Errorf("[Code: %d] Conflict: %s", v.Payload.Code, msg(v.Payload.Message))
	case *endpoint.GetAPINotFound:
		return i18n.Errorf("[Code: %d] get api error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *endpoint.GetAPIInternalServerError:
		return i18n.Errorf("[Code: %d] get api error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Create
	case *endpoint.AddAPIBadRequest:
		return i18n.Errorf("[Code: %d] create api error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *endpoint.AddAPIUnauthorized:
		return i18n.Errorf("[Code: %d] create api error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *endpoint.AddAPIInternalServerError:
		return i18n.Errorf("[Code: %d] create api error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Update
	case *endpoint.UpdateAPINotFound:
		return i18n.Errorf("[Code: %d] update api error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *endpoint.UpdateAPIBadRequest:
		return i18n.Errorf("[Code: %d] update api error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *endpoint.UpdateAPIInternalServerError:
		return i18n.Errorf("[Code: %d] update api error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Delete
	case *endpoint.DeleteAPIBadRequest:
		return i18n.Errorf("[Code: %d] delete api error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *endpoint.DeleteAPINotFound:
		return i18n.Errorf("[Code: %d] delete api error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *endpoint.DeleteAPIInternalServerError:
		return i18n.Errorf("[Code: %d] delete api error: %s", v.Payload.Code, msg(v.Payload.Message))

	// Policy
	// Add
	case *policy.AddPolicyConflict:
		return i18n.Errorf("[Code: %d] Add Policy conflict: %s", v.Payload.Code, msg(v.Payload.Message))
	case *policy.AddPolicyBadRequest:
		return i18n.Errorf("[Code: %d] Add Policy bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *policy.AddPolicyInternalServerError:
		return i18n.Errorf("[Code: %d] Add Policy internal server rrror: %s", v.Payload.Code, msg(v.Payload.Message))
	case *policy.AddPolicyDefault:
		return i18n.Errorf("[Code: %d] Add Policy error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Delete
	case *policy.DeletePolicyNotFound:
		return i18n.Errorf("[Code: %d] Delete Policy not found: %s", v.Payload.Code, msg(v.Payload.Message))
	case *policy.DeletePolicyBadRequest:
		return i18n.Errorf("[Code: %d] Delete Policy bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *policy.DeletePolicyInternalServerError:
		return i18n.Errorf("[Code: %d] Delete Policy internal server error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Get
	case *policy.GetPolicyNotFound:
		return i18n.Errorf("[Code: %d] Get Policy not found: %s", v.Payload.Code, msg(v.Payload.Message))
	case *policy.GetPoliciesDefault:
		return i18n.Errorf("[Code: %d] Get Policy error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *policy.GetPolicyBadRequest:
		return i18n.Errorf("[Code: %d] Get Policy bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	// Update
	case *policy.UpdatePolicyNotFound:
		return i18n.Errorf("[Code: %d] Update Policy not found: %s", v.Payload.Code, msg(v.Payload.Message))
	case *policy.UpdatePolicyInternalServerError:
		return i18n.Errorf("[Code: %d] Update Policy internal server error: %s", v.Payload.Code, msg(v.Payload.Message))

	// ServiceAccount
	// Add
	case *serviceaccount.AddServiceAccountConflict:
		return i18n.Errorf("[Code: %d] Add Service Account conflict: %s", v.Payload.Code, msg(v.Payload.Message))
	case *serviceaccount.AddServiceAccountBadRequest:
		return i18n.Errorf("[Code: %d] Add Service Account bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *serviceaccount.AddServiceAccountInternalServerError:
		return i18n.Errorf("[Code: %d] Add Service Account internal server rrror: %s", v.Payload.Code, msg(v.Payload.Message))
	case *serviceaccount.AddServiceAccountDefault:
		return i18n.Errorf("[Code: %d] Add Service Account error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Delete
	case *serviceaccount.DeleteServiceAccountNotFound:
		return i18n.Errorf("[Code: %d] Delete Service Account not found: %s", v.Payload.Code, msg(v.Payload.Message))
	case *serviceaccount.DeleteServiceAccountBadRequest:
		return i18n.Errorf("[Code: %d] Delete Service Account bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *serviceaccount.DeleteServiceAccountInternalServerError:
		return i18n.Errorf("[Code: %d] Delete Service Account internal server error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Get
	case *serviceaccount.GetServiceAccountNotFound:
		return i18n.Errorf("[Code: %d] Get Service Account not found: %s", v.Payload.Code, msg(v.Payload.Message))
	case *serviceaccount.GetServiceAccountsDefault:
		return i18n.Errorf("[Code: %d] Get Service Account error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *serviceaccount.GetServiceAccountBadRequest:
		return i18n.Errorf("[Code: %d] Get Service Account bad request: %s", v.Payload.Code, msg(v.Payload.Message))

	default:
		return i18n.Errorf("received unexpected error: %+v", v)
	}
}
