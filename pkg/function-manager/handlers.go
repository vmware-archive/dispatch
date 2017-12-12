///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-openapi/runtime"
	apiclient "github.com/go-openapi/runtime/client"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/function-manager/gen/models"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations"
	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/store"
	"github.com/vmware/dispatch/pkg/functions"
	imageclient "github.com/vmware/dispatch/pkg/image-manager/gen/client"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	imageclientimage "github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	imagemodels "github.com/vmware/dispatch/pkg/image-manager/gen/models"
	secretclient "github.com/vmware/dispatch/pkg/secret-store/gen/client"
	"github.com/vmware/dispatch/pkg/trace"
)

// FunctionManagerFlags are configuration flags for the function manager
var FunctionManagerFlags = struct {
	Config       string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile       string `long:"db-file" description:"Path to BoltDB file" default:"./db.bolt"`
	OrgID        string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
	ImageManager string `long:"image-manager" description:"Image manager endpoint" default:"localhost:8002"`
	SecretStore  string `long:"secret-store" description:"Secret store endpoint" default:"localhost:8003"`
	Faas         string `long:"faas" description:"FaaS implementation" default:"openfaas"`
}{}

func functionEntityToModel(f *Function) *models.Function {
	defer trace.Trace("functionEntityToModel")()
	var tags []*models.Tag
	for k, v := range f.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	return &models.Function{
		CreatedTime: f.CreatedTime.Unix(),
		Name:        swag.String(f.Name),
		ID:          strfmt.UUID(f.ID),
		Image:       swag.String(f.ImageName),
		Code:        swag.String(f.Code),
		Schema: &models.Schema{
			In:  f.Schema.In,
			Out: f.Schema.Out,
		},
		Secrets: f.Secrets,
		Tags:    tags,
	}
}

func functionListToModel(funcs []Function) []*models.Function {
	defer trace.Trace("functionListToModel")()
	body := make([]*models.Function, 0, len(funcs))
	for _, f := range funcs {
		body = append(body, functionEntityToModel(&f))
	}
	return body
}

func schemaModelToEntity(mSchema *models.Schema) (*Schema, error) {
	schema := new(Schema)
	if mSchema.In != nil {
		schema.In = new(spec.Schema)
		b, _ := json.Marshal(mSchema.In)
		if err := json.Unmarshal(b, schema.In); err != nil {
			return nil, errors.Wrap(err, "could not decode schema.in")
		}
	}
	if mSchema.Out != nil {
		schema.Out = new(spec.Schema)
		b, _ := json.Marshal(mSchema.Out)
		if err := json.Unmarshal(b, schema.Out); err != nil {
			return nil, errors.Wrap(err, "could not decode schema.out")
		}
	}
	return schema, nil
}

func functionModelOntoEntity(m *models.Function, e *Function) error {
	defer trace.Trace("functionModelOntoEntity")()
	schema, err := schemaModelToEntity(m.Schema)
	if err != nil {
		return err
	}
	main := "main"
	if m.Main != nil && *m.Main != "" {
		main = *m.Main
	}
	e.Code = *m.Code
	e.Main = main
	e.ImageName = *m.Image
	e.Tags = map[string]string{}
	for _, t := range m.Tags {
		e.Tags[t.Key] = t.Value
	}
	e.Schema = schema
	e.Secrets = m.Secrets
	return nil
}

func runModelToEntity(m *models.Run, f *Function) *FnRun {
	defer trace.Trace("runModelToEntity")()
	secrets := f.Secrets
	if m.Secrets != nil && len(m.Secrets) > 0 {
		secrets = m.Secrets
	}
	return &FnRun{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: FunctionManagerFlags.OrgID,
			Name:           uuid.NewV4().String(),
		},
		Blocking:     m.Blocking,
		Input:        m.Input,
		Secrets:      secrets,
		FunctionName: f.Name,
	}
}

func runEntityToModel(f *FnRun) *models.Run {
	defer trace.Trace("runEntityToModel")()
	m := models.Run{
		ExecutedTime: f.CreatedTime.Unix(),
		FinishedTime: f.ModifiedTime.Unix(),
		Name:         strfmt.UUID(f.Name),
		Blocking:     f.Blocking,
		Input:        f.Input,
		Output:       f.Output,
		Logs:         f.Logs,
		Secrets:      f.Secrets,
		FunctionName: f.FunctionName,
	}
	return &m
}

func runListToModel(runs []FnRun) []*models.Run {
	defer trace.Trace("runListToModel")()
	body := make([]*models.Run, 0, len(runs))
	for _, r := range runs {
		body = append(body, runEntityToModel(&r))
	}
	return body
}

type Handlers struct {
	FaaS         functions.FaaSDriver
	Runner       functions.Runner
	Store        entitystore.EntityStore
	ImgClient    ImageManager
	SecretClient *secretclient.SecretStore
}

type ImageManager interface {
	GetImageByName(*imageclientimage.GetImageByNameParams, runtime.ClientAuthInfoWriter) (*imageclientimage.GetImageByNameOK, error)
}

func ImageManagerClient() ImageManager {
	defer trace.Trace("ImageManagerClient")()
	transport := httptransport.New(FunctionManagerFlags.ImageManager, imageclient.DefaultBasePath, []string{"http"})
	return imageclient.New(transport, strfmt.Default).Image
}

func SecretStoreClient() *secretclient.SecretStore {
	defer trace.Trace("SecretStoreClient")()
	transport := httptransport.New(FunctionManagerFlags.SecretStore, secretclient.DefaultBasePath, []string{"http"})
	return secretclient.New(transport, strfmt.Default)
}

// ConfigureHandlers registers the function manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	defer trace.Trace("ConfigureHandlers")()
	a, ok := api.(*operations.FunctionManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		log.Printf("cookie auth: %s\n", token)
		return token, nil
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
	a.RunnerGetFunctionRunsHandler = fnrunner.GetFunctionRunsHandlerFunc(h.getFunctionRuns)
}

func (h *Handlers) addFunction(params fnstore.AddFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreAddFunctionHandler")()

	// get the auth cookie from the auth middleware "cookieAuth",
	// note this code is temporary and will be refactored
	cookie, ok := principal.(string)
	if !ok {
		return fnstore.NewAddFunctionUnauthorized().WithPayload(&models.Error{Message: swag.String("Invalid Cookie")})
	}

	e := &Function{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: FunctionManagerFlags.OrgID,
			Name:           *params.Body.Name,
		},
	}

	if err := functionModelOntoEntity(params.Body, e); err != nil {
		return fnstore.NewAddFunctionBadRequest().WithPayload(&models.Error{Message: swag.String(err.Error())})
	}
	image, err := h.getImage(e.ImageName, cookie)
	if err != nil {
		log.Errorf("Error when fetching image for function %s: %+v", e.Name, err)
		return fnstore.NewAddFunctionBadRequest().WithPayload(&models.Error{
			UserError: struct{}{},
			Code:      http.StatusBadRequest,
			Message:   swag.String(err.Error()),
		})
	}
	if err = h.FaaS.Create(e.Name, &functions.Exec{
		Code:     e.Code,
		Main:     e.Main,
		Image:    image.DockerURL,
		Language: string(image.Language),
	}); err != nil {
		log.Errorf("Driver error when creating a FaaS function: %+v", err)
		return fnstore.NewAddFunctionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when creating a Faas function"),
		})
	}
	if _, err := h.Store.Add(e); err != nil {
		log.Errorf("Store error when adding a new function %s: %+v", e.Name, err)
		return fnstore.NewAddFunctionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing a new function"),
		})
	}
	m := functionEntityToModel(e)
	return fnstore.NewAddFunctionOK().WithPayload(m)
}

func (h *Handlers) getFunction(params fnstore.GetFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreGetFunctionHandler")()
	e := new(Function)
	if err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, e); err != nil {
		log.Debugf("Error returned by h.Store.Get: ", err)
		log.Infof("Received GET for non-existent function %s", params.FunctionName)
		return fnstore.NewGetFunctionNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("function not found"),
		})
	}
	return fnstore.NewGetFunctionOK().WithPayload(functionEntityToModel(e))
}

func (h *Handlers) deleteFunction(params fnstore.DeleteFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreDeleteFunctionHandler")()
	e := new(Function)
	if err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, e); err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Received DELETE for non-existent function %s", params.FunctionName)
		return fnstore.NewDeleteFunctionNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("function not found"),
		})
	}
	if err := h.Store.Delete(FunctionManagerFlags.OrgID, params.FunctionName, e); err != nil {
		log.Errorf("Store error when deleting a function %s: %+v", params.FunctionName, err)
		return fnstore.NewDeleteFunctionBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String("error when deleting a function"),
		})
	}
	if err := h.FaaS.Delete(e.Name); err != nil {
		log.Errorf("Driver error when deleting a FaaS function: %+v", err)
		return fnstore.NewDeleteFunctionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting a FaaS function"),
		})
	}
	e.Delete = true
	m := functionEntityToModel(e)
	return fnstore.NewDeleteFunctionOK().WithPayload(m)
}

func (h *Handlers) getFunctions(params fnstore.GetFunctionsParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreGetFunctionsHandler")()
	funcs := []Function{}
	err := h.Store.List(FunctionManagerFlags.OrgID, nil, &funcs)
	if err != nil {
		log.Errorf("Store error when listing functions: %+v\n", err)
		return fnstore.NewGetFunctionsDefault(http.StatusInternalServerError).WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("error when listing functions"),
		})
	}
	return fnstore.NewGetFunctionsOK().WithPayload(functionListToModel(funcs))
}

func (h *Handlers) updateFunction(params fnstore.UpdateFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreUpdateFunctionHandler")()

	// get the auth cookie from the auth middleware "cookieAuth",
	// note this code is temporary and will be refactored
	cookie, ok := principal.(string)
	if !ok {
		log.Errorf("unauthorized: invalid cookie")
		return fnstore.NewAddFunctionUnauthorized().WithPayload(&models.Error{
			Code:    http.StatusUnauthorized,
			Message: swag.String("unauthorized: invalid cookie"),
		})
	}

	e := new(Function)
	err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, e)
	if err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Received update for non-existent function %s", params.FunctionName)
		return fnstore.NewDeleteFunctionNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("function not found"),
		})
	}

	if err := functionModelOntoEntity(params.Body, e); err != nil {
		return fnstore.NewUpdateFunctionBadRequest().WithPayload(&models.Error{
			UserError: struct{}{},
			Code:      http.StatusBadRequest,
			Message:   swag.String(err.Error()),
		})
	}
	image, err := h.getImage(e.ImageName, cookie)
	if err != nil {
		log.Errorf("Error when fetching image for function %s: %+v", e.Name, err)
		return fnstore.NewUpdateFunctionBadRequest().WithPayload(&models.Error{
			UserError: struct{}{},
			Code:      http.StatusBadRequest,
			Message:   swag.String(err.Error()),
		})
	}
	if err := h.FaaS.Create(e.Name, &functions.Exec{
		Code:     e.Code,
		Main:     e.Main,
		Image:    image.DockerURL,
		Language: string(image.Language),
	}); err != nil {
		log.Errorf("Driver error when creating a FaaS function: %+v", err)
		return fnstore.NewUpdateFunctionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when creating a FaaS function"),
		})
	}
	if _, err := h.Store.Update(e.Revision, e); err != nil {
		log.Errorf("Store error when updating function %s: %+v", params.FunctionName, err)
		return fnstore.NewUpdateFunctionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when updating a FaaS function"),
		})
	}
	m := functionEntityToModel(e)
	return fnstore.NewUpdateFunctionOK().WithPayload(m)
}

func (h *Handlers) runFunction(params fnrunner.RunFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("RunnerRunFunctionHandler")()

	// get the auth cookie from the auth middleware "cookieAuth",
	// note this code is temporary and will be refactored
	cookie, ok := principal.(string)
	if !ok {
		return fnstore.NewAddFunctionUnauthorized().WithPayload(&models.Error{Message: swag.String("Invalid Cookie")})
	}
	if params.Body == nil {
		return fnrunner.NewRunFunctionBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String("Bad Request: Invalid Payload"),
		})
	}
	log.Debugf("Execute a function with payload: %#v", *params.Body)

	e := new(Function)
	if err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, e); err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Trying to create run for non-existent function %s", params.FunctionName)
		return fnrunner.NewRunFunctionNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("function not found"),
		})
	}
	run := runModelToEntity(params.Body, e)
	if run.Blocking {
		ctx := functions.Context{}
		output, err := h.Runner.Run(&functions.Function{
			Context: ctx,
			Name:    e.Name,
			Schemas: &functions.Schemas{
				SchemaIn:  e.Schema.In,
				SchemaOut: e.Schema.Out,
			},
			Cookie:  cookie,
			Secrets: run.Secrets,
		}, run.Input)
		if err != nil {
			if userError, ok := err.(functions.UserError); ok {
				return fnrunner.NewRunFunctionBadRequest().WithPayload(&models.Error{
					Code:      http.StatusBadRequest,
					Message:   swag.String(err.Error()),
					UserError: userError.AsUserErrorObject(),
				})
			}
			if functionError, ok := err.(functions.FunctionError); ok {
				return fnrunner.NewRunFunctionBadGateway().WithPayload(&models.Error{
					Code:          http.StatusBadRequest,
					Message:       swag.String(err.Error()),
					FunctionError: functionError.AsFunctionErrorObject(),
				})
			}
			log.Errorf("Driver error when running function %s: %+v", e.Name, err)
			return fnrunner.NewRunFunctionInternalServerError().WithPayload(&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when running a function"),
			})
		}
		run.Output = output
		run.Logs = ctx.Logs()
		_, err = h.Store.Add(run)
		return fnrunner.NewRunFunctionOK().WithPayload(runEntityToModel(run))
	}
	if _, err := h.Store.Add(run); err != nil {
		log.Errorf("Store error when adding new function run %s: %+v", run.Name, err)
		return fnrunner.NewRunFunctionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing the new function"),
		})
	}
	// TODO call the function asynchronously
	return fnrunner.NewRunFunctionAccepted().WithPayload(runEntityToModel(run))
}

func (h *Handlers) getRun(params fnrunner.GetRunParams, principal interface{}) middleware.Responder {
	defer trace.Trace("RunnerGetRunHandler")()
	run := FnRun{}
	err := h.Store.Get(FunctionManagerFlags.OrgID, params.RunName.String(), &run)
	if err != nil || run.FunctionName != params.FunctionName {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Get run failed for function %s and run %s", params.FunctionName, params.RunName.String())
		return fnrunner.NewGetRunNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("internal server error when getting a function run"),
		})
	}
	return fnrunner.NewGetRunOK().WithPayload(runEntityToModel(&run))
}

func (h *Handlers) getRuns(params fnrunner.GetRunsParams, principal interface{}) middleware.Responder {
	defer trace.Trace("RunnerGetRunsHandler")()
	var filter entitystore.Filter
	var runs []FnRun
	if err := h.Store.List(FunctionManagerFlags.OrgID, filter, &runs); err != nil {
		log.Errorf("Store error when listing runs: %+v", err)
		return fnrunner.NewGetRunsNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("error when listing function runs"),
		})
	}
	return fnrunner.NewGetRunsOK().WithPayload(runListToModel(runs))
}

func (h *Handlers) getFunctionRuns(params fnrunner.GetFunctionRunsParams, principal interface{}) middleware.Responder {
	defer trace.Trace("RunnerGetRunsHandler")()
	f := new(Function)
	var filter entitystore.Filter
	if params.FunctionName != "" {
		if err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, f); err != nil {
			log.Debugf("Error returned by h.Store.Get: %+v", err)
			log.Infof("Trying to list runs for non-existent function: %s", params.FunctionName)
			return fnrunner.NewGetRunsNotFound().WithPayload(&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("internal server error when getting the function"),
			})
		}
		filter = func(e entitystore.Entity) bool {
			run, ok := e.(*FnRun)
			return ok && run.FunctionName == f.Name
		}
	}
	var runs []FnRun
	if err := h.Store.List(FunctionManagerFlags.OrgID, filter, &runs); err != nil {
		log.Errorf("Store error when listing runs for function %s: %+v", params.FunctionName, err)
		return fnrunner.NewGetRunsNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("error when listing function runs"),
		})
	}
	return fnrunner.NewGetRunsOK().WithPayload(runListToModel(runs))
}

func (h *Handlers) getImage(imageName, cookie string) (*imagemodels.Image, error) {

	apiKeyAuth := apiclient.APIKeyAuth("cookie", "header", cookie)
	resp, err := h.ImgClient.GetImageByName(
		&image.GetImageByNameParams{
			ImageName: imageName,
			Context:   context.Background(),
		}, apiKeyAuth)
	if err == nil {
		return resp.Payload, nil
	}
	return nil, errors.Wrapf(err, "failed to get image: '%s'", imageName)
}
