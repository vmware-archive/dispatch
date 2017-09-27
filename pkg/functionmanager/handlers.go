///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package functionmanager

import (
	"context"
	"fmt"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	uuid "github.com/satori/go.uuid"

	entitystore "gitlab.eng.vmware.com/serverless/serverless/pkg/entity-store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations"
	fnrunner "gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations/runner"
	fnstore "gitlab.eng.vmware.com/serverless/serverless/pkg/functionmanager/gen/restapi/operations/store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions"
	imageclient "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/image"
)

// FunctionManagerFlags are configuration flags for the function manager
var FunctionManagerFlags = struct {
	Config       string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile       string `long:"db-file" description:"Path to BoltDB file" default:"./db.bolt"`
	OrgID        string `long:"organization" description:"(temporary) Static organization id" default:"serverless"`
	ImageManager string `long:"image-manager" description:"Image manager endpoint" default:"localhost:8002"`
}{}

func functionEntityToModel(f *Function) *models.Function {
	var tags []*models.Tag
	for k, v := range f.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}
	schema := models.Schema(f.Schema)
	m := models.Function{
		CreatedTime: f.CreatedTime.Unix(),
		Name:        swag.String(f.Name),
		ID:          strfmt.UUID(f.ID),
		Image:       swag.String(f.ImageName),
		Code:        swag.String(f.Code),
		Schema:      &schema,

		Tags: tags,
	}
	return &m
}

func functionListToModel(funcs []Function) []*models.Function {
	body := make([]*models.Function, 0, len(funcs))
	for _, f := range funcs {
		body = append(body, functionEntityToModel(&f))
	}
	return body
}

func functionModelToEntity(m *models.Function) *Function {
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	e := Function{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: FunctionManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
		},
		Code:      *m.Code,
		ImageName: *m.Image,
		Schema:    Schema(*m.Schema),
	}
	e.ID = string(m.ID)
	return &e
}

func runModelToEntity(m *models.Run) *FnRun {
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
	transport := httptransport.New(FunctionManagerFlags.ImageManager, imageclient.DefaultBasePath, []string{"http"})
	return imageclient.New(transport, strfmt.Default)
}

// ConfigureHandlers registers the function manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI, store entitystore.EntityStore) {

	a, ok := api.(*operations.FunctionManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.StoreAddFunctionHandler = fnstore.AddFunctionHandlerFunc(func(params fnstore.AddFunctionParams) middleware.Responder {
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
			fmt.Println(err)
		}
		e := functionModelToEntity(functionRequest)
		if err := h.FaaS.Create(e.Name, &functions.Exec{
			Code:  e.Code,
			Main:  "main",    // TODO add "main" field to Function type in the swagger spec
			Image: dockerURL, // TODO get the docker image name by e.ImageName from image-manager
		}); err != nil {
			fnstore.NewAddFunctionMethodNotAllowed() // TODO respond with appropriate error
		}
		_, err = store.Add(e)
		if err != nil {
			return fnstore.NewAddFunctionMethodNotAllowed()
		}
		m := functionEntityToModel(e)
		return fnstore.NewAddFunctionAccepted().WithPayload(m)
	})

	a.StoreGetFunctionByNameHandler = fnstore.GetFunctionByNameHandlerFunc(func(params fnstore.GetFunctionByNameParams) middleware.Responder {
		e := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			return fnstore.NewGetFunctionByNameNotFound()
		}
		m := functionEntityToModel(&e)
		return fnstore.NewGetFunctionByNameOK().WithPayload(m)
	})

	a.StoreDeleteFunctionByNameHandler = fnstore.DeleteFunctionByNameHandlerFunc(func(params fnstore.DeleteFunctionByNameParams) middleware.Responder {
		e := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			return fnstore.NewDeleteFunctionByNameNotFound()
		}
		err = store.Delete(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			return fnstore.NewDeleteFunctionByNameBadRequest()
		}
		return fnstore.NewDeleteFunctionByNameNoContent()
	})

	a.StoreGetFunctionsHandler = fnstore.GetFunctionsHandlerFunc(func(params fnstore.GetFunctionsParams) middleware.Responder {
		funcs := []Function{}
		err := store.List(FunctionManagerFlags.OrgID, nil, funcs)
		if err != nil {
			return fnstore.NewGetFunctionsDefault(500)
		}
		return fnstore.NewGetFunctionsOK().WithPayload(functionListToModel(funcs))
	})

	a.StoreUpdateFunctionByNameHandler = fnstore.UpdateFunctionByNameHandlerFunc(func(params fnstore.UpdateFunctionByNameParams) middleware.Responder {
		e := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			return fnstore.NewDeleteFunctionByNameNotFound()
		}
		e.Code = *params.Body.Code
		tags := make(map[string]string)
		for _, t := range params.Body.Tags {
			tags[t.Key] = t.Value
		}
		e.Tags = tags
		e.Schema = Schema(*params.Body.Schema)
		_, err = store.Update(e.Revision, &e)
		if err != nil {
			return fnstore.NewUpdateFunctionByNameBadRequest()
		}
		m := functionEntityToModel(&e)
		return fnstore.NewGetFunctionByNameOK().WithPayload(m)
	})

	a.RunnerRunFunctionHandler = fnrunner.RunFunctionHandlerFunc(func(params fnrunner.RunFunctionParams) middleware.Responder {
		e := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &e)
		if err != nil {
			return fnrunner.NewRunFunctionNotFound()
		}
		run := runModelToEntity(params.Body)
		run.FunctionName = e.Name
		if run.Blocking {
			output, err := h.Runner.Run(&functions.Function{
				Name:    e.Name,
				Schemas: &functions.Schemas{}, // TODO put the schemas from the stored function here
			}, run.Input.(map[string]interface{}))
			if err != nil {
				fmt.Println(err)
				return fnrunner.NewRunFunctionInternalServerError() // TODO proper error
			}
			run.Output = output
			_, err = store.Add(run)
			return fnrunner.NewRunFunctionOK().WithPayload(runEntityToModel(run))
		}
		_, err = store.Add(run)
		if err != nil {
			return fnrunner.NewRunFunctionInternalServerError()
		}
		if run.Blocking {
			return fnrunner.NewRunFunctionOK().WithPayload(runEntityToModel(run))
		}

		return fnrunner.NewRunFunctionAccepted().WithPayload(runEntityToModel(run))
	})

	a.RunnerGetRunByNameHandler = fnrunner.GetRunByNameHandlerFunc(func(params fnrunner.GetRunByNameParams) middleware.Responder {
		run := FnRun{}
		err := store.Get(FunctionManagerFlags.OrgID, params.RunName.String(), &run)
		if err != nil || run.FunctionName != params.FunctionName {
			return fnrunner.NewGetRunByNameNotFound()
		}
		return fnrunner.NewGetRunByNameOK().WithPayload(runEntityToModel(&run))
	})

	a.RunnerGetRunsHandler = fnrunner.GetRunsHandlerFunc(func(params fnrunner.GetRunsParams) middleware.Responder {
		f := Function{}
		err := store.Get(FunctionManagerFlags.OrgID, params.FunctionName, &f)
		if err != nil {
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
			return fnrunner.NewGetRunsNotFound()
		}
		return fnrunner.NewGetRunsOK().WithPayload(runListToModel(runs))
	})
}
