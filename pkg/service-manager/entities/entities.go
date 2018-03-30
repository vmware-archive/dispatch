///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entities

import (
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/service-manager/flags"
	"github.com/vmware/dispatch/pkg/service-manager/gen/models"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// Broker represents a service broker (which implements OSBAPI).
type Broker struct {
	entitystore.BaseEntity
	URL string `json:"url"`
}

// Schema represents contract for the three service operations (Create, Update, and Bind).
type Schema struct {
	Create *spec.Schema `json:"create,omitempty"`
	Update *spec.Schema `json:"update,omitempty"`
	Bind   *spec.Schema `json:"bind,omitempty"`
}

// ServicePlan represents a plan or flavor of a service.  Different plans may have different schemas as well.
type ServicePlan struct {
	entitystore.BaseEntity
	Description string      `json:"description"`
	PlanID      string      `json:"planID"`
	Schema      Schema      `json:"schema"`
	Metadata    interface{} `json:"metadata"`
	Free        bool        `json:"free"`
	Bindable    bool        `json:"bindable"`
}

// ServiceClass represents an available service type.  The service plans are associated with the service class.
type ServiceClass struct {
	entitystore.BaseEntity
	Description string        `json:"description"`
	ServiceID   string        `json:"serviceID"`
	Broker      string        `json:"broker"`
	Bindable    bool          `json:"bindable"`
	Plans       []ServicePlan `json:"plans"`
}

// ServiceBinding represents a binding or connection to the service.  Generally this is in the form of credentials
// which can be made available to a function.
type ServiceBinding struct {
	entitystore.BaseEntity
	ServiceInstance  string      `json:"serviceInstance"`
	Parameters       interface{} `json:"parameters"`
	SecretParameters []string    `json:"secretParameters"`
	BindingID        string      `json:"bindingID"`
	BindingSecret    string      `json:"bindingSecret"`
}

// ServiceInstance represents a provisioned service.
type ServiceInstance struct {
	entitystore.BaseEntity
	ServiceClass     string          `json:"serviceClass"`
	ServicePlan      string          `json:"servicePlan"`
	Namespace        string          `json:"namespace"`
	Parameters       interface{}     `json:"parameters"`
	SecretParameters []string        `json:"secretParameters"`
	InstanceID       string          `json:"instanceID"`
	Bind             bool            `json:"bind"`
	Binding          *ServiceBinding `json:"binding,omitempty"`
}

var statusMap = map[models.Status]entitystore.Status{
	models.StatusCREATING:    entitystore.StatusCREATING,
	models.StatusDELETED:     entitystore.StatusDELETED,
	models.StatusERROR:       entitystore.StatusERROR,
	models.StatusINITIALIZED: entitystore.StatusINITIALIZED,
	models.StatusREADY:       entitystore.StatusREADY,
}
var reverseStatusMap = make(map[entitystore.Status]models.Status)

// InitializeStatusMap initializes the status mapping
func InitializeStatusMap() {
	defer trace.Trace("")()
	for k, v := range statusMap {
		reverseStatusMap[v] = k
	}
}

// ServiceClassEntityToModel translates the ServiceClass entity representation (DB) to the model representation (API).
func ServiceClassEntityToModel(e *ServiceClass) *models.ServiceClass {
	defer trace.Trace("")()
	var tags []*models.Tag
	for k, v := range e.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}

	var plans []*models.ServicePlan
	for _, plan := range e.Plans {
		plans = append(plans, &models.ServicePlan{
			ID:          strfmt.UUID(plan.ID),
			Name:        plan.Name,
			Kind:        utils.ServicePlanKind,
			Description: plan.Description,
			Metadata:    plan.Metadata,
			Free:        plan.Free,
			Bindable:    plan.Bindable,
			Schema: &models.ServicePlanSchema{
				Create: plan.Schema.Create,
				Update: plan.Schema.Update,
				Bind:   plan.Schema.Bind,
			},
		})
	}

	m := models.ServiceClass{
		CreatedTime: e.CreatedTime.Unix(),
		ID:          strfmt.UUID(e.ID),
		Name:        swag.String(e.Name),
		Kind:        utils.ServiceClassKind,
		Status:      reverseStatusMap[e.Status],
		Tags:        tags,
		Reason:      e.Reason,
		Broker:      swag.String(e.Broker),
		Bindable:    e.Bindable,
		Plans:       plans,
	}
	return &m
}

// ServiceClassModelToEntity translates the ServiceClass model representation (API) to the entity representation (DB).
func ServiceClassModelToEntity(m *models.ServiceClass) *ServiceClass {
	defer trace.Trace("")()
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	e := ServiceClass{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: flags.ServiceManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
			Status:         statusMap[m.Status],
			Reason:         m.Reason,
		},
	}
	return &e
}

// ServiceInstanceEntityToModel translates the ServiceInstance entity representation (DB) to the model representation
// (API).  Notice that the ServiceBinding is includeded as the API does not have a separate binding endpoint.
// Services are always created with a single binding.
func ServiceInstanceEntityToModel(e *ServiceInstance, b *ServiceBinding) *models.ServiceInstance {
	defer trace.Trace("")()
	var tags []*models.Tag
	for k, v := range e.Tags {
		tags = append(tags, &models.Tag{Key: k, Value: v})
	}

	m := models.ServiceInstance{
		CreatedTime:      e.CreatedTime.Unix(),
		ID:               strfmt.UUID(e.ID),
		Name:             swag.String(e.Name),
		Kind:             utils.ServiceInstanceKind,
		Status:           reverseStatusMap[e.Status],
		Tags:             tags,
		Reason:           e.Reason,
		ServiceClass:     swag.String(e.ServiceClass),
		ServicePlan:      swag.String(e.ServicePlan),
		Parameters:       e.Parameters,
		SecretParameters: e.SecretParameters,
	}
	if b != nil {
		m.Binding = &models.ServiceBinding{
			Status:           reverseStatusMap[b.Status],
			Parameters:       b.Parameters,
			SecretParameters: b.SecretParameters,
			BindingSecret:    b.BindingSecret,
		}
	}
	return &m
}

// ServiceInstanceModelToEntity translates the ServiceInstance model representation (API) to the entity representation
// (DB).  Notice that the ServiceBinding is includeded as the API does not have a separate binding endpoint.
// Services are always created with a single binding.
func ServiceInstanceModelToEntity(m *models.ServiceInstance) (*ServiceInstance, *ServiceBinding) {
	defer trace.Trace("")()
	tags := make(map[string]string)
	for _, t := range m.Tags {
		tags[t.Key] = t.Value
	}
	e := ServiceInstance{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: flags.ServiceManagerFlags.OrgID,
			Name:           *m.Name,
			Tags:           tags,
		},
		ServiceClass:     *m.ServiceClass,
		ServicePlan:      *m.ServicePlan,
		Parameters:       m.Parameters,
		SecretParameters: m.SecretParameters,
	}
	b := ServiceBinding{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: flags.ServiceManagerFlags.OrgID,
			Name:           *m.Name,
		},
	}
	if m.Binding != nil {
		b.Parameters = m.Binding.Parameters
		b.SecretParameters = m.Binding.SecretParameters
		b.BindingSecret = m.Binding.BindingSecret
	}
	return &e, &b
}
