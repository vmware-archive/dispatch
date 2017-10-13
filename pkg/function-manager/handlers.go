///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package functionmanager

import (
	"context"
	"encoding/json"

	apiclient "github.com/go-openapi/runtime/client"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/restapi/operations"
	fnrunner "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/restapi/operations/runner"
	fnstore "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/restapi/operations/store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
	imageclient "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/image"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/trace"
)

// FunctionManagerFlags are configuration flags for the function manager
var FunctionManagerFlags = struct {
	Config       string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile       string `long:"db-file" description:"Path to BoltDB file" default:"./db.bolt"`
	OrgID        string `long:"organization" description:"(temporary) Static organization id" default:"serverless"`
	ImageManager string `long:"image-manager" description:"Image manager endpoint" default:"localhost:8002"`
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

		Tags: tags,
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
	return nil
}

func runModelToEntity(m *models.Run) *FnRun {
	defer trace.Trace("runModelToEntity")()
	return &FnRun{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: FunctionManagerFlags.OrgID,
			Name:           uuid.NewV4().String(),
		},
		Blocking: m.Blocking,
		Input:    m.Input,
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
	FaaS      functions.FaaSDriver
	Runner    functions.Runner
	Store     entitystore.EntityStore
	ImgClient *imageclient.ImageManager
}

func ImageManagerClient() *imageclient.ImageManager {
	defer trace.Trace("ImageManagerClient")()
	transport := httptransport.New(FunctionManagerFlags.ImageManager, imageclient.DefaultBasePath, []string{"http"})
	return imageclient.New(transport, strfmt.Default)
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
	if err := h.FaaS.Create(e.Name, &functions.Exec{
		Code:  e.Code,
		Main:  e.Main,
		Image: h.getDockerImage(e.ImageName, cookie),
	}); err != nil {
		log.Errorf("Driver error when creating a FaaS function: %+v", err)
		return fnstore.NewAddFunctionInternalServerError().WithPayload(&models.Error{Message: swag.String(err.Error())})
	}
	if _, err := h.Store.Add(e); err != nil {
		log.Errorf("Store error when adding a new function %s: %+v", e.Name, err)
		return fnstore.NewAddFunctionInternalServerError().WithPayload(&models.Error{Message: swag.String(err.Error())})
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
		return fnstore.NewGetFunctionNotFound()
	}
	return fnstore.NewGetFunctionOK().WithPayload(functionEntityToModel(e))
}

func (h *Handlers) deleteFunction(params fnstore.DeleteFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreDeleteFunctionHandler")()
	e := new(Function)
	if err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, e); err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Received DELETE for non-existent function %s", params.FunctionName)
		return fnstore.NewDeleteFunctionNotFound()
	}
	if err := h.Store.Delete(FunctionManagerFlags.OrgID, params.FunctionName, e); err != nil {
		log.Errorf("Store error when deleting a function %s: %+v", params.FunctionName, err)
		return fnstore.NewDeleteFunctionBadRequest()
	}
	if err := h.FaaS.Delete(e.Name); err != nil {
		log.Errorf("Driver error when deleting a FaaS function: %+v", err)
		return fnstore.NewDeleteFunctionInternalServerError().WithPayload(&models.Error{
			Message: swag.String(err.Error()),
		})
	}
	return fnstore.NewDeleteFunctionNoContent()
}

func (h *Handlers) getFunctions(params fnstore.GetFunctionsParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreGetFunctionsHandler")()
	funcs := []Function{}
	err := h.Store.List(FunctionManagerFlags.OrgID, nil, &funcs)
	if err != nil {
		log.Errorf("Store error when listing functions: %+v", err)
		return fnstore.NewGetFunctionsDefault(500)
	}
	return fnstore.NewGetFunctionsOK().WithPayload(functionListToModel(funcs))
}

func (h *Handlers) updateFunction(params fnstore.UpdateFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreUpdateFunctionHandler")()

	// get the auth cookie from the auth middleware "cookieAuth",
	// note this code is temporary and will be refactored
	cookie, ok := principal.(string)
	if !ok {
		return fnstore.NewAddFunctionUnauthorized().WithPayload(&models.Error{Message: swag.String("Invalid Cookie")})
	}

	e := new(Function)
	err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, e)
	if err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Received update for non-existent function %s", params.FunctionName)
		return fnstore.NewDeleteFunctionNotFound()
	}

	if err := functionModelOntoEntity(params.Body, e); err != nil {
		return fnstore.NewUpdateFunctionBadRequest().WithPayload(&models.Error{
			UserError: struct{}{},
			Message:   swag.String(err.Error()),
		})
	}
	if err := h.FaaS.Create(e.Name, &functions.Exec{
		Code:  e.Code,
		Main:  e.Main,
		Image: h.getDockerImage(e.ImageName, cookie),
	}); err != nil {
		log.Errorf("Driver error when creating a FaaS function: %+v", err)
		return fnstore.NewUpdateFunctionInternalServerError().WithPayload(&models.Error{
			Message: swag.String(err.Error()),
		})
	}
	if _, err := h.Store.Update(e.Revision, e); err != nil {
		log.Errorf("Store error when updating function %s: %+v", params.FunctionName, err)
		return fnstore.NewUpdateFunctionInternalServerError().WithPayload(&models.Error{
			Message: swag.String(err.Error()),
		})
	}
	m := functionEntityToModel(e)
	return fnstore.NewUpdateFunctionOK().WithPayload(m)
}

func (h *Handlers) runFunction(params fnrunner.RunFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("RunnerRunFunctionHandler")()
	e := new(Function)
	if err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, e); err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Trying to create run for non-existent function %s", params.FunctionName)
		return fnrunner.NewRunFunctionNotFound()
	}
	run := runModelToEntity(params.Body)
	run.FunctionName = e.Name
	if run.Blocking {
		output, err := h.Runner.Run(&functions.Function{
			Name: e.Name,
			Schemas: &functions.Schemas{
				SchemaIn:  e.Schema.In,
				SchemaOut: e.Schema.Out,
			},
		}, run.Input.(map[string]interface{}))
		if err != nil {
			if userError, ok := err.(functions.UserError); ok {
				return fnrunner.NewRunFunctionBadRequest().WithPayload(&models.Error{
					Message:   swag.String(err.Error()),
					UserError: userError.AsUserErrorObject(),
				})
			}
			if functionError, ok := err.(functions.FunctionError); ok {
				return fnrunner.NewRunFunctionBadGateway().WithPayload(&models.Error{
					Message:       swag.String(err.Error()),
					FunctionError: functionError.AsFunctionErrorObject(),
				})
			}
			log.Errorf("Driver error when running function %s: %+v", e.Name, err)
			return fnrunner.NewRunFunctionInternalServerError().WithPayload(&models.Error{Message: swag.String(err.Error())})
		}
		run.Output = output
		_, err = h.Store.Add(run)
		return fnrunner.NewRunFunctionOK().WithPayload(runEntityToModel(run))
	}
	if _, err := h.Store.Add(run); err != nil {
		log.Errorf("Store error when adding new function run %s: %+v", run.Name, err)
		return fnrunner.NewRunFunctionInternalServerError()
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
		return fnrunner.NewGetRunNotFound()
	}
	return fnrunner.NewGetRunOK().WithPayload(runEntityToModel(&run))
}

func (h *Handlers) getRuns(params fnrunner.GetRunsParams, principal interface{}) middleware.Responder {
	defer trace.Trace("RunnerGetRunsHandler")()
	f := Function{}
	err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &f)
	if err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Trying to list runs for non-existent function: %s", params.FunctionName)
		return fnrunner.NewGetRunsNotFound()
	}
	filter := func(e entitystore.Entity) bool {
		if run, ok := e.(*FnRun); ok {
			return run.FunctionName == f.Name
		}
		return false
	}
	runs := []FnRun{}
	err = h.Store.List(FunctionManagerFlags.OrgID, entitystore.Filter(filter), runs)
	if err != nil {
		log.Errorf("Store error when listing runs for function %s: %+v", params.FunctionName, err)
		return fnrunner.NewGetRunsNotFound()
	}
	return fnrunner.NewGetRunsOK().WithPayload(runListToModel(runs))
}

func (h *Handlers) getDockerImage(imageName, cookie string) string {

	apiKeyAuth := apiclient.APIKeyAuth("cookie", "header", cookie)
	if resp, err := h.ImgClient.Image.GetImageByName(&image.GetImageByNameParams{
		ImageName: imageName,
		Context:   context.Background(),
	}, apiKeyAuth); err == nil {
		// TODO (bjung) fix this!!!
		return resp.Payload.DockerURL
	} else {
		log.Errorf("%+v", errors.Wrapf(err, "failed to get docker image URL, imageName: '%s'", imageName))
	}
	return imageName
}
