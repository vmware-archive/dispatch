///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/functions/backend"
	"github.com/vmware/dispatch/pkg/functions/gen/restapi/operations"
	fnrunner "github.com/vmware/dispatch/pkg/functions/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/functions/gen/restapi/operations/store"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

//Handlers interface declares methods needed to implement functions API
type Handlers interface {
	addFunction(params fnstore.AddFunctionParams) middleware.Responder
	getFunction(params fnstore.GetFunctionParams) middleware.Responder
	deleteFunction(params fnstore.DeleteFunctionParams) middleware.Responder
	getFunctions(params fnstore.GetFunctionsParams) middleware.Responder
	updateFunction(params fnstore.UpdateFunctionParams) middleware.Responder
	runFunction(params fnrunner.RunFunctionParams) middleware.Responder
	getRun(params fnrunner.GetRunParams) middleware.Responder
	getRuns(params fnrunner.GetRunsParams) middleware.Responder
}

// ConfigureHandlers registers the function manager knHandlers to the API
func ConfigureHandlers(api middleware.RoutableAPI, h Handlers) {
	a, ok := api.(*operations.FunctionsAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.Logger = log.Printf
	a.StoreAddFunctionHandler = fnstore.AddFunctionHandlerFunc(h.addFunction)
	a.StoreGetFunctionHandler = fnstore.GetFunctionHandlerFunc(h.getFunction)
	a.StoreDeleteFunctionHandler = fnstore.DeleteFunctionHandlerFunc(h.deleteFunction)
	a.StoreGetFunctionsHandler = fnstore.GetFunctionsHandlerFunc(h.getFunctions)
	a.StoreUpdateFunctionHandler = fnstore.UpdateFunctionHandlerFunc(h.updateFunction)
	a.RunnerRunFunctionHandler = fnrunner.RunFunctionHandlerFunc(h.runFunction)
	a.RunnerGetRunHandler = fnrunner.GetRunHandlerFunc(h.getRun)
	a.RunnerGetRunsHandler = fnrunner.GetRunsHandlerFunc(h.getRuns)
}

type defaultHandlers struct {
	backend        backend.Backend
	httpClient     *http.Client
	namespace      string
	imageRegistry  string
	sourceRootPath string
	imagesClient   client.ImagesClient
}

// NewHandlers is the constructor for the function manager API knHandlers
func NewHandlers(kubeconfPath, namespace, imageRegistry, sourceRootPath string, imagesClient client.ImagesClient) Handlers {
	return &defaultHandlers{
		backend:        backend.Knative(kubeconfPath),
		httpClient:     &http.Client{},
		namespace:      namespace,
		imageRegistry:  imageRegistry,
		sourceRootPath: sourceRootPath,
		imagesClient:   imagesClient,
	}
}

func (h *defaultHandlers) addFunction(params fnstore.AddFunctionParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	function := params.Body
	utils.AdjustMeta(&function.Meta, dapi.Meta{Org: org, Project: project})

	sourceID := uuid.NewV4().String()
	sourceDir := fmt.Sprintf("%s/%s/%s", h.sourceRootPath, org, project)
	sourcePath := fmt.Sprintf("%s/%s.tgz", sourceDir, sourceID)
	// TODO (bjung): need to support object store as well (far easier for local development)
	sourceURL := fmt.Sprintf("file://%s", sourcePath)

	img, err := h.imagesClient.GetImage(ctx, org, function.Image)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "fetching image for function"))
		return fnstore.NewAddFunctionDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function", function.Meta.Name),
		})
	}
	function.ImageURL = img.ImageDestination
	log.Debugf("fetched image url %s for image %s and function %s", img.ImageDestination, function.Image, function.Name)

	os.MkdirAll(sourceDir, 0700)
	err = ioutil.WriteFile(sourcePath, function.Source, 0600)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "writing function source"))
		return fnstore.NewAddFunctionDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function", function.Meta.Name),
		})
	}
	// Once saved, unset source as we don't need it anymore
	// What is the best way to transmit source?  Probably not through JSON like
	// we are doing.
	function.Source = nil
	function.SourceURL = sourceURL
	function.FunctionImageURL = fmt.Sprintf("%s/%s", h.imageRegistry, sourceID)

	createdFunction, err := h.backend.Add(ctx, function)
	if err != nil {
		if _, ok := err.(backend.AlreadyExists); ok {
			return fnstore.NewAddFunctionConflict().WithPayload(&dapi.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("function", function.Meta.Name),
			})
		}
		log.Errorf("%+v", errors.Wrap(err, "creating a function"))
		return fnstore.NewAddFunctionDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function", function.Meta.Name),
		})
	}

	return fnstore.NewAddFunctionCreated().WithPayload(createdFunction)
}

func (h *defaultHandlers) getFunction(params fnstore.GetFunctionParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	name := params.FunctionName
	log.Debugf("getting function %s in %s:%s", name, org, project)

	function, err := h.backend.Get(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		if _, ok := err.(backend.NotFound); ok {
			return fnstore.NewGetFunctionNotFound().WithPayload(&dapi.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("function", name),
			})
		}
		errors.Wrapf(err, "getting function '%s'", name)
	}

	return fnstore.NewGetFunctionOK().WithPayload(function)
}

func (h *defaultHandlers) deleteFunction(params fnstore.DeleteFunctionParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	name := params.FunctionName

	err := h.backend.Delete(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		if _, ok := err.(backend.NotFound); ok {
			return fnstore.NewDeleteFunctionNotFound().WithPayload(&dapi.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("function", name),
			})
		}
		errors.Wrapf(err, "deleting function '%s'", name)
	}

	return fnstore.NewDeleteFunctionOK()
}

func (h *defaultHandlers) getFunctions(params fnstore.GetFunctionsParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject
	log.Debugf("getting functions in %s:%s", org, project)
	functions, err := h.backend.List(ctx, &dapi.Meta{Org: org, Project: project})
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "listing functions"))
		return fnstore.NewGetFunctionsDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String(err.Error()),
		})
	}

	return fnstore.NewGetFunctionsOK().WithPayload(functions)
}

func (h *defaultHandlers) updateFunction(params fnstore.UpdateFunctionParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	function := params.Body
	utils.AdjustMeta(&function.Meta, dapi.Meta{Org: org, Project: project})

	updatedFunction, err := h.backend.Update(ctx, function)
	if err != nil {
		if _, ok := err.(backend.NotFound); ok {
			return fnstore.NewUpdateFunctionNotFound().WithPayload(&dapi.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("function", function.Meta.Name),
			})
		}
		log.Errorf("%+v", errors.Wrap(err, "updating a function"))
		return fnstore.NewUpdateFunctionDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function", function.Meta.Name),
		})
	}

	return fnstore.NewUpdateFunctionOK().WithPayload(updatedFunction)
}

func (h *defaultHandlers) runFunction(params fnrunner.RunFunctionParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject
	name := *params.FunctionName

	contentType := params.Body.HTTPContext["Content-Type"].(string)
	accept := params.Body.HTTPContext["Accept"].(string)
	inBytes := params.Body.InputBytes

	runEndpoint, err := h.backend.RunEndpoint(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		if _, ok := err.(backend.NotFound); ok {
			return fnrunner.NewRunFunctionNotFound().WithPayload(&dapi.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("function", name),
			})
		}
		errors.Wrapf(err, "getting function '%s'", name)
	}

	req, err := http.NewRequest("POST", runEndpoint, bytes.NewReader(inBytes))
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "building http request"))
		return fnrunner.NewRunFunctionDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("building http request to run function", name),
		})
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", accept)
	// TODO: Add Dispatch context via header (X-Dispatch-Context)

	response, err := h.httpClient.Do(req)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "performing http request"))
		return fnrunner.NewRunFunctionDefault(502).WithPayload(&dapi.Error{
			Code:    http.StatusBadGateway,
			Message: utils.ErrorMsgInternalError("performing http request to run function", name),
		})
	}
	defer response.Body.Close()

	outContentType := response.Header.Get("Content-Type")
	outBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "reading http response body"))
		return fnrunner.NewRunFunctionDefault(502).WithPayload(&dapi.Error{
			Code:    http.StatusBadGateway,
			Message: utils.ErrorMsgInternalError("reading http response body running function", name),
		})
	}

	run := &dapi.Run{
		HTTPContext: map[string]interface{}{"Content-Type": outContentType},
		OutputBytes: outBytes,
	}
	return fnrunner.NewRunFunctionOK().WithPayload(run)
}

func (*defaultHandlers) getRun(params fnrunner.GetRunParams) middleware.Responder {
	panic("implement me")
}

func (*defaultHandlers) getRuns(params fnrunner.GetRunsParams) middleware.Responder {
	panic("implement me")
}
