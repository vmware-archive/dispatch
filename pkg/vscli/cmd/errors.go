///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	runner "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/client/runner"
	function "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/client/store"
	baseimage "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/base_image"
	image "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/image"
	secret "gitlab.eng.vmware.com/serverless/serverless/pkg/secret-store/gen/client/secret"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

func msg(m *string) string {
	if m == nil {
		return ""
	}
	return *m
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
	case *baseimage.AddBaseImageDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Delete
	case *baseimage.DeleteBaseImageByNameBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *baseimage.DeleteBaseImageByNameNotFound:
		p := params.(*baseimage.GetBaseImageByNameParams)
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
	case *image.AddImageDefault:
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
	case *function.AddFunctionUnauthorized:
		return i18n.Errorf("[Code: %d] Unauthorized: %s", v.Payload.Code, msg(v.Payload.Message))
	case *function.AddFunctionInternalServerError:
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
		return i18n.Errorf("[Code: %d] Function execution not found: %s", v.Payload.Code, p.FunctionName)
	case *runner.RunFunctionInternalServerError:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *runner.RunFunctionBadGateway:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// List
	case *runner.GetRunsNotFound:
		p := params.(*runner.GetRunsParams)
		return i18n.Errorf("[Code: %d] Function executions not found: %s", v.Payload.Code, p.FunctionName)
	// Secret
	// Get
	case *secret.GetSecretNotFound:
		return i18n.Errorf("[Code: %d] get Secret not found: %s", v.Payload.Code, msg(v.Payload.Message))
	case *secret.GetSecretDefault:
		return i18n.Errorf("[Code: %d] get Secret error: %s", v.Payload.Code, msg(v.Payload.Message))
	case *secret.GetSecretsDefault:
		return i18n.Errorf("[Code: %d] get Secret error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Create
	case *secret.AddSecretDefault:
		return i18n.Errorf("[Code: %d] create Secret error: %s", v.Payload.Code, msg(v.Payload.Message))
	default:
		return i18n.Errorf("received unexpected error: %+v", v)
	}
}
