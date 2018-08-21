///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	knclientset "github.com/knative/eventing/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	driverapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/drivers"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
	"github.com/vmware/dispatch/pkg/utils/knaming"
)

const (
	// EventDriverType is the event driver type name
	EventDriverType = "EventDriverType"
)

// KnativeHandlers is a base struct for event manager drivers API handlers.
type KnativeHandlers struct {
	eventingClient *knclientset.Clientset
	secretsClient  client.SecretsClient
}

// EventDriverTypeName returns k8s API name of the Dispatch function
func EventDriverTypeName(meta v1.Meta) string {
	return knaming.GetKnName("EventDriverType", meta)
}

// NewKnativeHandlers Creates new instance of driver handlers
func NewKnativeHandlers(secretsClient client.SecretsClient) *KnativeHandlers {
	k8sconfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(errors.Wrap(err, "error configuring k8s API client"))
	}
	return &KnativeHandlers{
		eventingClient: knclientset.NewForConfigOrDie(k8sconfig),
		secretsClient:  secretsClient,
	}
}

// ConfigureHandlers configures API handlers for driver endpoints
func (h *KnativeHandlers) ConfigureHandlers(api middleware.RoutableAPI) {
	a, ok := api.(*operations.EventManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.DriversAddDriverTypeHandler = driverapi.AddDriverTypeHandlerFunc(h.addDriverType)
	a.DriversGetDriverTypeHandler = driverapi.GetDriverTypeHandlerFunc(h.getDriverType)
	a.DriversGetDriverTypesHandler = driverapi.GetDriverTypesHandlerFunc(h.getDriverTypes)
	a.DriversDeleteDriverTypeHandler = driverapi.DeleteDriverTypeHandlerFunc(h.deleteDriverType)
}

func (h *KnativeHandlers) addDriverType(params driverapi.AddDriverTypeParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		return driverapi.NewAddDriverTypeBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("invalid driver type payload: %s", err)),
		})
	}

	driverType := *params.Body

	driverType.Meta = v1.Meta{
		Org:     params.XDispatchOrg,
		Project: *params.XDispatchProject,
		Name:    *driverType.Name,
	}

	eventSource := FromDriverType(&driverType)

	_, err := h.eventingClient.FeedsV1alpha1().EventSources(driverType.Meta.Org).Create(eventSource)

	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			return driverapi.NewAddDriverTypeConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("event driver type", *driverType.Name),
			})
		}
		log.Errorf("error when adding a new driver type %s: %+v", *driverType.Name, err)
		return driverapi.NewAddDriverTypeDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("event driver type", *driverType.Name),
		})
	}

	return driverapi.NewAddDriverTypeCreated().WithPayload(&driverType)
}

func (h *KnativeHandlers) getDriverType(params driverapi.GetDriverTypeParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	dispatchMeta := v1.Meta{
		Org:     params.XDispatchOrg,
		Project: *params.XDispatchProject,
		Name:    params.DriverTypeName,
	}

	name := EventDriverTypeName(dispatchMeta)

	eventSource, err := h.eventingClient.FeedsV1alpha1().EventSources(dispatchMeta.Org).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Warnf("Received GET for non-existent driver type %s", params.DriverTypeName)
		log.Debugf("error when getting driver type: %+v", err)
		return driverapi.NewGetDriverNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("event driver type", params.DriverTypeName),
			})
	}
	return driverapi.NewGetDriverTypeOK().WithPayload(ToDriverType(eventSource))
}

func (h *KnativeHandlers) getDriverTypes(params driverapi.GetDriverTypesParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	org := params.XDispatchOrg
	project := *params.XDispatchProject

	labelSelector := knaming.ToLabelSelector(
		map[string]string{
			knaming.OrgLabel:     org,
			knaming.ProjectLabel: project,
		},
	)
	eventSources, err := h.eventingClient.FeedsV1alpha1().EventSources(org).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Errorf("error when listing driver types: %+v", err)
		return driverapi.NewGetDriverTypesDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting driver types"),
			})
	}
	var driverTypeModels []*v1.EventDriverType
	for _, eventSource := range eventSources.Items {
		driverTypeModels = append(driverTypeModels, ToDriverType(&eventSource))
	}

	return driverapi.NewGetDriverTypesOK().WithPayload(driverTypeModels)
}

func (h *KnativeHandlers) deleteDriverType(params driverapi.DeleteDriverTypeParams, principal interface{}) middleware.Responder {
	span, _ := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	dispatchMeta := v1.Meta{
		Org:     params.XDispatchOrg,
		Project: *params.XDispatchProject,
		Name:    params.DriverTypeName,
	}

	name := EventDriverTypeName(dispatchMeta)

	eventSource, err := h.eventingClient.FeedsV1alpha1().EventSources(dispatchMeta.Org).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Warnf("Received GET for non-existent driver type %s", params.DriverTypeName)
		log.Debugf("error when getting driver type: %+v", err)
		return driverapi.NewGetDriverNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("event driver type", params.DriverTypeName),
			})
	}

	if err = h.eventingClient.FeedsV1alpha1().EventSources(dispatchMeta.Org).Delete(name, nil); err != nil {
		if k8sErrors.IsNotFound(err) {
			return driverapi.NewDeleteDriverTypeNotFound().WithPayload(
				&v1.Error{
					Code:    http.StatusNotFound,
					Message: utils.ErrorMsgNotFound("event driver type", params.DriverTypeName),
				})
		}
		log.Errorf("error when deleting the event driver type %s: %+v", params.DriverTypeName, err)
		return driverapi.NewDeleteDriverTypeDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("event driver type", params.DriverTypeName),
		})
	}
	// TODO: return code should be 204
	return driverapi.NewDeleteDriverTypeOK().WithPayload(ToDriverType(eventSource))
}
