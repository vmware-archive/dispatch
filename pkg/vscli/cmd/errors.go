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
	case *function.AddFunctionInternalServerError:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Delete
	case *function.DeleteFunctionByNameBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *function.DeleteFunctionByNameNotFound:
		p := params.(*function.DeleteFunctionByNameParams)
		return i18n.Errorf("[Code: %d] Function not found: %s", v.Payload.Code, p.FunctionName)
	// Get
	case *function.GetFunctionByNameBadRequest:
		return i18n.Errorf("[Code: %d] Bad request: %s", v.Payload.Code, msg(v.Payload.Message))
	case *function.GetFunctionByNameNotFound:
		p := params.(*function.GetFunctionByNameParams)
		return i18n.Errorf("[Code: %d] Function not found: %s", v.Payload.Code, p.FunctionName)
	// List
	case *function.GetFunctionsDefault:
		return i18n.Errorf("[Code: %d] Error: %s", v.Payload.Code, msg(v.Payload.Message))
	// Runner
	// Get
	case *runner.GetRunByNameNotFound:
		p := params.(*runner.GetRunByNameParams)
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
	}
	return err
}
