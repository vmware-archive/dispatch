///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functions

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	minio "github.com/minio/minio-go"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	dapi "github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/functions/backend"
	"github.com/vmware/dispatch/pkg/functions/config"
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
	backend       backend.Backend
	httpClient    *http.Client
	namespace     string
	imageRegistry string
	storageConfig *config.StorageConfig
	imagesClient  client.ImagesClient
}

// NewHandlers is the constructor for the function manager API knHandlers
func NewHandlers(kubeconfPath, namespace, imageRegistry, ingressGateway, buildImage string, storageConfig *config.StorageConfig, imagesClient client.ImagesClient) Handlers {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &defaultHandlers{
		backend:       backend.Knative(kubeconfPath, ingressGateway, buildImage, storageConfig),
		httpClient:    &http.Client{Transport: tr},
		namespace:     namespace,
		imageRegistry: imageRegistry,
		imagesClient:  imagesClient,
		storageConfig: storageConfig,
	}
}

func (h *defaultHandlers) writeSource(sourceID, org, project string, source []byte) (*url.URL, error) {
	name := fmt.Sprintf("%s.tgz", sourceID)
	switch h.storageConfig.Storage {
	case config.Minio:
		minioClient, err := minio.New(
			h.storageConfig.Minio.MinioAddress, h.storageConfig.Minio.Username, h.storageConfig.Minio.Password, false)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create minio client")
		}
		contentType := "application/tar+gz"
		bucketName := fmt.Sprintf("%s-%s", org, project)

		err = minioClient.MakeBucket(bucketName, string(h.storageConfig.Minio.Location))
		if err != nil {
			_, err := minioClient.BucketExists(bucketName)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create minio bucket")
			}
		}

		tmpfile, err := ioutil.TempFile("", name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create temp file")
		}
		defer os.Remove(tmpfile.Name())

		err = ioutil.WriteFile(tmpfile.Name(), source, 0600)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to write temp file")
		}
		_, err = minioClient.FPutObject(bucketName, name, tmpfile.Name(), minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			return nil, errors.Wrapf(err, "error putting file to minio bucket")
		}
		return url.Parse(fmt.Sprintf("minio://%s/%s/%s", h.storageConfig.Minio.MinioAddress, bucketName, name))
	case config.File:
		sourceDir := fmt.Sprintf("%s/%s/%s", h.storageConfig.File.SourceRootPath, org, project)
		sourcePath := fmt.Sprintf("%s/%s", sourceDir, name)
		os.MkdirAll(sourceDir, 0700)
		err := ioutil.WriteFile(sourcePath, source, 0600)
		if err != nil {
			return nil, errors.Wrapf(err, "error writing file")
		}
		return url.Parse(fmt.Sprintf("file://%s/%s", sourcePath, name))
	default:
		return nil, fmt.Errorf("unknown source URL scheme: %s", h.storageConfig.Storage)
	}
}

func (h *defaultHandlers) addFunction(params fnstore.AddFunctionParams) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := h.namespace
	project := *params.XDispatchProject

	function := params.Body
	utils.AdjustMeta(&function.Meta, dapi.Meta{Org: org, Project: project})

	img, err := h.imagesClient.GetImage(ctx, org, function.Image)
	if err != nil {
		if err, ok := err.(client.Error); ok {
			return fnstore.NewAddFunctionDefault(err.Code()).WithPayload(&dapi.Error{
				Code:    int64(err.Code()),
				Message: swag.String(err.Message()),
			})
		}
		log.Errorf("%+v", errors.Wrap(err, "fetching image for function"))
		return fnstore.NewAddFunctionDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function", function.Meta.Name),
		})
	}
	function.ImageURL = img.ImageURL
	log.Debugf("fetched image url %s for image %s and function %s", img.ImageURL, function.Image, function.Name)

	sourceID := uuid.NewV4().String()
	u, err := h.writeSource(sourceID, org, project, function.Source)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "writing function source"))
		return fnstore.NewAddFunctionDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("function", function.Meta.Name),
		})
	}
	sourceURL := u.String()
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

	log.Debugf("running function %s:%s:%s", org, project, name)
	runHost, runEndpoint, err := h.backend.RunEndpoint(ctx, &dapi.Meta{Name: name, Org: org, Project: project})
	if err != nil {
		if _, ok := err.(backend.NotFound); ok {
			return fnrunner.NewRunFunctionNotFound().WithPayload(&dapi.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("function", name),
			})
		}
		errors.Wrapf(err, "getting function '%s'", name)
	}
	req, err := http.NewRequest("POST", runHost, bytes.NewReader(inBytes))
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "building http request"))
		return fnrunner.NewRunFunctionDefault(500).WithPayload(&dapi.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("building http request to run function", name),
		})
	}
	req.Host = runEndpoint
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", accept)
	// TODO: Make timeout configurable
	h.httpClient.Timeout = 60 * time.Second
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
	// Technically, this shouldn't happen... but it will.
	if response.StatusCode == http.StatusNotFound {
		return fnrunner.NewRunFunctionNotFound().WithPayload(&dapi.Error{
			Code:    http.StatusNotFound,
			Message: utils.ErrorMsgInternalError("function not found", name),
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
