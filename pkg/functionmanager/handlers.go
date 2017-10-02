///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package functionmanager

import (
	"context"
	"encoding/json"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations"
	fnrunner "gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations/runner"
	fnstore "gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations/store"
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

func functionModelToEntity(m *models.Function) (*Function, error) {
	defer trace.Trace("functionModelToEntity")()
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	schema, err := schemaModelToEntity(m.Schema)
	if err != nil {
		return nil, err
	}
	main := "main"
	if m.Main != nil && *m.Main != "" {
		main = *m.Main
	}
	e := Function{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: FunctionManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
		},
		Code:      *m.Code,
		Main:      main,
		ImageName: *m.Image,
		Schema:    schema,
	}
	e.ID = string(m.ID)
	return &e, nil
}

func runModelToEntity(m *models.Run) *FnRun {
	defer trace.Trace("runModelToEntity")()
	e := FnRun{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: FunctionManagerFlags.OrgID,
			Name:           uuid.NewV4().String(),
		},
		Blocking: m.Blocking,
		Input:    m.Input,
	}
	return &e
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
	FaaS   functions.FaaSDriver
	Runner functions.Runner
}

func imageManagerClient() *imageclient.ImageManager {
	defer trace.Trace("imageManagerClient")()
	transport := httptransport.New(FunctionManagerFlags.ImageManager, imageclient.DefaultBasePath, []string{"http"})
	return imageclient.New(transport, strfmt.Default)
}

// ConfigureHandlers registers the function manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI, store entitystore.EntityStore) {
	defer trace.Trace("ConfigureHandlers")()
	a, ok := api.(*operations.FunctionManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}
	a.Logger = log.Printf
	a.StoreAddFunctionHandler = fnstore.AddFunctionHandlerFunc(func(params fnstore.AddFunctionParams) middleware.Responder {
		defer trace.Trace("StoreAddFunctionHandler")()
		functionRequest := params.Body

		imgClient := imageManagerClient()
		resp, err := imgClient.Image.GetImageByName(&image.GetImageByNameParams{
			ImageName: *functionRequest.Image,
			Context:   context.Background(),
		})
		dockerURL := *functionRequest.Image
		if err == nil {
			// TODO (bjung) fix this!!!
			dockerURL = resp.Payload.DockerURL
		} else {
			log.Errorln(err)
		}
		e, err := functionModelToEntity(functionRequest)
		if err != nil {
			return fnstore.NewAddFunctionBadRequest().WithPayload(err)
		}
		if err := h.FaaS.Create(e.Name, &functions.Exec{
			Code:  e.Code,
			Main:  e.Main,
			Image: dockerURL,
		}); err != nil {
			log.Errorf("Driver error when creating a FaaS function: %+v", err)
			return fnstore.NewAddFunctionInternalServerError().WithPayload(err)
		}
		if _, err := store.Add(e); err != nil {
			log.Errorf("Store error when adding a new function %s: %+v", e.Name, err)
			return fnstore.NewAddFunctionInternalServerError().WithPayload(err)
		}
		m := functionEntityToModel(e)
		return fnstore.NewAddFunctionOK().WithPayload(m)
	})

	a.StoreGetFunctionByNameHandler = fnstore.GetFunctionByNameHandlerFunc(func(params fnstore.GetFunctionByNameParams) middleware.Responder {
		defer trace.Trace("StoreGetFunctionByNameHandler")()
		e := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			log.Debugf("Error returned by store.Get: ", err)
			log.Infof("Received GET for non-existent function %s", params.FunctionName)
			return fnstore.NewGetFunctionByNameNotFound()
		}
		m := functionEntityToModel(&e)
		return fnstore.NewGetFunctionByNameOK().WithPayload(m)
	})

	a.StoreDeleteFunctionByNameHandler = fnstore.DeleteFunctionByNameHandlerFunc(func(params fnstore.DeleteFunctionByNameParams) middleware.Responder {
		defer trace.Trace("StoreDeleteFunctionByNameHandler")()
		e := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			log.Debugf("Error returned by store.Get: %+v", err)
			log.Infof("Received DELETE for non-existent function %s", params.FunctionName)
			return fnstore.NewDeleteFunctionByNameNotFound()
		}
		err = store.Delete(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			log.Errorf("Store error when deleting a function %s: %+v", params.FunctionName, err)
			return fnstore.NewDeleteFunctionByNameBadRequest()
		}
		return fnstore.NewDeleteFunctionByNameNoContent()
	})

	a.StoreGetFunctionsHandler = fnstore.GetFunctionsHandlerFunc(func(params fnstore.GetFunctionsParams) middleware.Responder {
		defer trace.Trace("StoreGetFunctionsHandler")()
		funcs := []Function{}
		err := store.List(FunctionManagerFlags.OrgID, nil, &funcs)
		if err != nil {
			log.Errorf("Store error when listing functions: %+v", err)
			return fnstore.NewGetFunctionsDefault(500)
		}
		return fnstore.NewGetFunctionsOK().WithPayload(functionListToModel(funcs))
	})

	a.StoreUpdateFunctionByNameHandler = fnstore.UpdateFunctionByNameHandlerFunc(func(params fnstore.UpdateFunctionByNameParams) middleware.Responder {
		defer trace.Trace("StoreUpdateFunctionByNameHandler")()
		e := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			log.Debugf("Error returned by store.Get: %+v", err)
			log.Infof("Received update for non-existent function %s", params.FunctionName)
			return fnstore.NewDeleteFunctionByNameNotFound()
		}
		e.Code = *params.Body.Code
		tags := make(map[string]string)
		for _, t := range params.Body.Tags {
			tags[t.Key] = t.Value
		}
		e.Tags = tags
		e.Schema, err = schemaModelToEntity(params.Body.Schema)
		if err != nil {
			return fnstore.NewUpdateFunctionByNameBadRequest()
		}
		_, err = store.Update(e.Revision, &e)
		if err != nil {
			log.Errorf("Store error when updating function %s: %+v", params.FunctionName, err)
			return fnstore.NewUpdateFunctionByNameBadRequest()
		}
		m := functionEntityToModel(&e)
		return fnstore.NewGetFunctionByNameOK().WithPayload(m)
	})

	a.RunnerRunFunctionHandler = fnrunner.RunFunctionHandlerFunc(func(params fnrunner.RunFunctionParams) middleware.Responder {
		defer trace.Trace("RunnerRunFunctionHandler")()
		e := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			log.Debugf("Error returned by store.Get: %+v", err)
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
				if err, ok := err.(functions.UserError); ok {
					return fnrunner.NewRunFunctionBadRequest().WithPayload(err.AsUserErrorObject())
				}
				if err, ok := err.(functions.FunctionError); ok {
					return fnrunner.NewRunFunctionBadGateway().WithPayload(err.AsFunctionErrorObject())
				}
				log.Errorf("Driver error when running function %s: %+v", e.Name, err)
				return fnrunner.NewRunFunctionInternalServerError().WithPayload(err)
			}
			run.Output = output
			_, err = store.Add(run)
			return fnrunner.NewRunFunctionOK().WithPayload(runEntityToModel(run))
		}
		_, err = store.Add(run)
		if err != nil {
			log.Errorf("Store error when adding new function run %s: %+v", run.Name, err)
			return fnrunner.NewRunFunctionInternalServerError()
		}
		// TODO call the function asynchronously
		return fnrunner.NewRunFunctionAccepted().WithPayload(runEntityToModel(run))
	})

	a.RunnerGetRunByNameHandler = fnrunner.GetRunByNameHandlerFunc(func(params fnrunner.GetRunByNameParams) middleware.Responder {
		defer trace.Trace("RunnerGetRunByNameHandler")()
		run := FnRun{}
		err := store.Get(FunctionManagerFlags.OrgID, params.RunName.String(), &run)
		if err != nil || run.FunctionName != params.FunctionName {
			log.Debugf("Error returned by store.Get: %+v", err)
			log.Infof("Get run failed for function %s and run %s", params.FunctionName, params.RunName.String())
			return fnrunner.NewGetRunByNameNotFound()
		}
		return fnrunner.NewGetRunByNameOK().WithPayload(runEntityToModel(&run))
	})

	a.RunnerGetRunsHandler = fnrunner.GetRunsHandlerFunc(func(params fnrunner.GetRunsParams) middleware.Responder {
		defer trace.Trace("RunnerGetRunsHandler")()
		f := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &f)
		if err != nil {
			log.Debugf("Error returned by store.Get: %+v", err)
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
		err = store.List(FunctionManagerFlags.OrgID, entitystore.Filter(filter), runs)
		if err != nil {
			log.Errorf("Store error when listing runs for function %s: %+v", params.FunctionName, err)
			return fnrunner.NewGetRunsNotFound()
		}
		return fnrunner.NewGetRunsOK().WithPayload(runListToModel(runs))
	})
}
