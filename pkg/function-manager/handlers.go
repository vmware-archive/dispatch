///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package functionmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	eventmodels "github.com/vmware/dispatch/pkg/event-manager/gen/models"
	"github.com/vmware/dispatch/pkg/event-manager/helpers"
	"github.com/vmware/dispatch/pkg/function-manager/gen/models"
	"github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations"
	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/runner"
	fnstore "github.com/vmware/dispatch/pkg/function-manager/gen/restapi/operations/store"
	"github.com/vmware/dispatch/pkg/functions"
	imageclient "github.com/vmware/dispatch/pkg/image-manager/gen/client"
	imageclientimage "github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	imagemodels "github.com/vmware/dispatch/pkg/image-manager/gen/models"
	secretclient "github.com/vmware/dispatch/pkg/secret-store/gen/client"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// FunctionManagerFlags are configuration flags for the function manager
var FunctionManagerFlags = struct {
	Config           string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile           string `long:"db-file" description:"Backend DB URL/Path" default:"./db.bolt"`
	DbBackend        string `long:"db-backend" description:"Backend DB Name" default:"boltdb"`
	DbUser           string `long:"db-username" description:"Backend DB Username" default:"dispatch"`
	DbPassword       string `long:"db-password" description:"Backend DB Password" default:"dispatch"`
	DbDatabase       string `long:"db-database" description:"Backend DB Name" default:"dispatch"`
	OrgID            string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
	ImageManager     string `long:"image-manager" description:"Image manager endpoint" default:"localhost:8002"`
	SecretStore      string `long:"secret-store" description:"Secret store endpoint" default:"localhost:8003"`
	K8sConfig        string `long:"kubeconfig" description:"Path to kubernetes config file" default:""`
	FileImageManager string `long:"file-image-manager" description:"Path to file containing images (useful for testing)"`
}{}

func functionEntityToModel(f *functions.Function) *models.Function {
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
		Status:  models.Status(f.Status),
	}
}

func functionListToModel(funcs []*functions.Function) []*models.Function {
	defer trace.Trace("functionListToModel")()
	body := make([]*models.Function, 0, len(funcs))
	for _, f := range funcs {
		body = append(body, functionEntityToModel(f))
	}
	return body
}

func schemaModelToEntity(mSchema *models.Schema) (*functions.Schema, error) {
	schema := new(functions.Schema)
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

func functionModelOntoEntity(m *models.Function, e *functions.Function) error {
	defer trace.Trace("functionModelOntoEntity")()

	e.BaseEntity = entitystore.BaseEntity{
		OrganizationID: FunctionManagerFlags.OrgID,
		Name:           *m.Name,
	}
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

func runModelToEntity(m *models.Run, f *functions.Function) *functions.FnRun {
	defer trace.Trace("runModelToEntity")()
	secrets := f.Secrets
	if secrets == nil {
		secrets = m.Secrets
	} else {
		secrets = append(secrets, m.Secrets...)
	}
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	var waitChan chan struct{}
	if m.Blocking {
		waitChan = make(chan struct{})
	}
	return &functions.FnRun{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: FunctionManagerFlags.OrgID,
			Name:           uuid.NewV4().String(),
			Status:         f.Status,
			Reason:         f.Reason,
			Tags:           tags,
		},
		Blocking:     m.Blocking,
		Input:        m.Input,
		Secrets:      secrets,
		FunctionName: f.Name,
		FunctionID:   f.ID,
		Event:        helpers.CloudEventFromSwagger((*eventmodels.CloudEvent)(m.Event)),
		WaitChan:     waitChan,
	}
}

func runEntityToModel(f *functions.FnRun) *models.Run {
	defer trace.Trace("runEntityToModel")()
	tags := []*models.Tag{}
	for k, v := range f.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	return &models.Run{
		ExecutedTime: f.CreatedTime.Unix(),
		FinishedTime: f.FinishedTime.Unix(),
		Name:         strfmt.UUID(f.Name),
		Blocking:     f.Blocking,
		Input:        f.Input,
		Output:       f.Output,
		Logs:         f.Logs,
		Secrets:      f.Secrets,
		FunctionName: f.FunctionName,
		FunctionID:   f.FunctionID,
		Status:       models.Status(f.Status),
		Event:        (*models.CloudEvent)(helpers.CloudEventToSwagger(f.Event)),
		Reason:       f.Reason,
		Tags:         tags,
	}
}

func runListToModel(runs []*functions.FnRun) []*models.Run {
	defer trace.Trace("runListToModel")()
	body := make([]*models.Run, 0, len(runs))
	for _, r := range runs {
		body = append(body, runEntityToModel(r))
	}
	return body
}

// Handlers is the API handler for function manager
type Handlers struct {
	Watcher controller.Watcher

	Store entitystore.EntityStore
}

// NewHandlers is the contstructor for the function manager API handlers
func NewHandlers(watcher controller.Watcher, store entitystore.EntityStore) *Handlers {
	return &Handlers{
		Watcher: watcher,
		Store:   store,
	}
}

// ImageManager is an interface to the image manager service
type ImageManager interface {
	GetImageByName(*imageclientimage.GetImageByNameParams, runtime.ClientAuthInfoWriter) (*imageclientimage.GetImageByNameOK, error)
}

// ImageManagerClient creates an ImageManager
func ImageManagerClient() ImageManager {
	defer trace.Trace("ImageManagerClient")()
	transport := httptransport.New(FunctionManagerFlags.ImageManager, imageclient.DefaultBasePath, []string{"http"})
	return imageclient.New(transport, strfmt.Default).Image
}

// FileImageManager is an ImageManager which is backed by a static map of images
type FileImageManager struct {
	Images map[string]*imagemodels.Image
}

// GetImageByName returns an image based queried by name
func (m *FileImageManager) GetImageByName(params *imageclientimage.GetImageByNameParams, writer runtime.ClientAuthInfoWriter) (*imageclientimage.GetImageByNameOK, error) {
	image, ok := m.Images[params.ImageName]
	if ok {
		return &imageclientimage.GetImageByNameOK{
			Payload: image,
		}, nil
	}
	return nil, fmt.Errorf("Missing image %s", params.ImageName)
}

// FileImageManagerClient returns a FileImageManager after populating the map with a JSON file
func FileImageManagerClient() ImageManager {
	defer trace.Trace("")()
	b, err := ioutil.ReadFile(FunctionManagerFlags.FileImageManager)
	if err != nil {
		panic(fmt.Sprintf("Failed to read image file %s", FunctionManagerFlags.FileImageManager))
	}
	images := make(map[string]*imagemodels.Image)
	json.Unmarshal(b, &images)
	return &FileImageManager{
		Images: images,
	}
}

// SecretStoreClient returns a client to the secret store
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

	e := &functions.Function{}
	if err := functionModelOntoEntity(params.Body, e); err != nil {
		return fnstore.NewAddFunctionBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(err.Error()),
		})
	}

	e.Status = entitystore.StatusINITIALIZED
	log.Debugf("trying to add entity to store")
	log.Debugf("entity org=%s, name=%s, id=%s, status=%s", e.OrganizationID, e.Name, e.ID, e.Status)
	if _, err := h.Store.Add(e); err != nil {
		log.Errorf("Store error when adding a new function %s: %+v", e.Name, err)
		return fnstore.NewAddFunctionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing a new function"),
		})
	}

	h.Watcher.OnAction(e)

	return fnstore.NewAddFunctionOK().WithPayload(functionEntityToModel(e))
}

func (h *Handlers) getFunction(params fnstore.GetFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreGetFunctionHandler")()
	e := new(functions.Function)

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnstore.NewGetFunctionBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	if err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, opts, e); err != nil {
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
	e := new(functions.Function)

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnstore.NewDeleteFunctionBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	if err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, opts, e); err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Received DELETE for non-existent function %s", params.FunctionName)
		return fnstore.NewDeleteFunctionNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("function not found"),
		})
	}

	e.Status = entitystore.StatusDELETING

	log.Debugf("trying to delete the entity from store")
	log.Debugf("entity org=%s, name=%s, id=%s, status=%s", e.OrganizationID, e.Name, e.ID, e.Status)
	if _, err := h.Store.Update(e.Revision, e); err != nil {
		log.Errorf("Store error when deleting a function %s: %+v", params.FunctionName, err)
		return fnstore.NewDeleteFunctionBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String("error when deleting a function"),
		})
	}
	h.Watcher.OnAction(e)
	m := functionEntityToModel(e)
	return fnstore.NewDeleteFunctionOK().WithPayload(m)
}

func (h *Handlers) getFunctions(params fnstore.GetFunctionsParams, principal interface{}) middleware.Responder {
	defer trace.Trace("StoreGetFunctionsHandler")()

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnstore.NewGetFunctionsBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	var funcs []*functions.Function
	err = h.Store.List(FunctionManagerFlags.OrgID, opts, &funcs)
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

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnstore.NewUpdateFunctionBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	e := new(functions.Function)
	err = h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, opts, e)
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

	e.Status = entitystore.StatusUPDATING

	if _, err := h.Store.Update(e.Revision, e); err != nil {
		log.Errorf("Store error when updating function %s: %+v", params.FunctionName, err)
		return fnstore.NewUpdateFunctionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when updating a FaaS function"),
		})
	}

	h.Watcher.OnAction(e)

	m := functionEntityToModel(e)
	return fnstore.NewUpdateFunctionOK().WithPayload(m)
}

func (h *Handlers) runFunction(params fnrunner.RunFunctionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("RunnerRunFunctionHandler")()
	var err error

	if params.Body == nil {
		return fnrunner.NewRunFunctionBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String("Bad Request: Invalid Payload"),
		})
	}
	log.Debugf("Execute a function with payload: %#v", *params.Body)

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnrunner.NewRunFunctionBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	f := new(functions.Function)
	if err := h.Store.Get(FunctionManagerFlags.OrgID, params.FunctionName, opts, f); err != nil {
		log.Debugf("Error returned by h.Store.Get: %+v", err)
		log.Infof("Trying to create run for non-existent function %s", params.FunctionName)
		return fnrunner.NewRunFunctionNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("function not found"),
		})
	}

	if f.Status != entitystore.StatusREADY {
		return fnrunner.NewRunFunctionNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("function is not READY"),
		})
	}

	run := runModelToEntity(params.Body, f)

	run.Status = entitystore.StatusINITIALIZED

	if _, err := h.Store.Add(run); err != nil {
		log.Errorf("Store error when adding new function run %s: %+v", run.Name, err)
		return fnrunner.NewRunFunctionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing the new function"),
		})
	}

	h.Watcher.OnAction(run)

	if run.Blocking {
		run.Wait()
		return fnrunner.NewRunFunctionOK().WithPayload(runEntityToModel(run))
	}

	return fnrunner.NewRunFunctionAccepted().WithPayload(runEntityToModel(run))
}

func (h *Handlers) getRun(params fnrunner.GetRunParams, principal interface{}) middleware.Responder {
	defer trace.Trace("RunnerGetRunHandler")()
	run := functions.FnRun{}

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnrunner.NewGetRunBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	err = h.Store.Get(FunctionManagerFlags.OrgID, params.RunName.String(), opts, &run)
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
	var runs []*functions.FnRun

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnrunner.NewGetRunsBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	if err = h.Store.List(FunctionManagerFlags.OrgID, opts, &runs); err != nil {
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

	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything().Add(
			entitystore.FilterStat{
				Scope:   entitystore.FilterScopeExtra,
				Subject: "FunctionName",
				Verb:    entitystore.FilterVerbEqual,
				Object:  params.FunctionName,
			}),
	}

	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return fnrunner.NewGetFunctionRunsBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	var runs []*functions.FnRun
	if err := h.Store.List(FunctionManagerFlags.OrgID, opts, &runs); err != nil {
		log.Errorf("Store error when listing runs for function %s: %+v", params.FunctionName, err)
		return fnrunner.NewGetFunctionRunsNotFound().WithPayload(&models.Error{
			Code:    http.StatusNotFound,
			Message: swag.String("error when listing function runs"),
		})
	}
	return fnrunner.NewGetFunctionRunsOK().WithPayload(runListToModel(runs))
}
