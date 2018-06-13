///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"fmt"
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/entity-store"
	organizationOperations "github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/organization"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

func organizationModelToEntity(m *v1.Organization) *Organization {
	e := Organization{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           *m.Name,
		},
	}
	return &e
}

func organizationEntityToModel(e *Organization) *v1.Organization {
	m := v1.Organization{
		ID:           strfmt.UUID(e.ID),
		Name:         swag.String(e.Name),
		Kind:         utils.OrganizationKind,
		Status:       v1.Status(e.Status),
		CreatedTime:  e.CreatedTime.Unix(),
		ModifiedTime: e.ModifiedTime.Unix(),
	}
	return &m
}

func (h *Handlers) getOrganizations(params organizationOperations.GetOrganizationsParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var organizations []*Organization

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	err := h.store.List(ctx, IdentityManagerFlags.OrgID, opts, &organizations)
	if err != nil {
		log.Errorf("store error when listing organizations: %+v", err)
		return organizationOperations.NewGetOrganizationsDefault(500).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting organizations"),
			})
	}
	var organizationModels []*v1.Organization
	for _, organization := range organizations {
		organizationModels = append(organizationModels, organizationEntityToModel(organization))
	}
	return organizationOperations.NewGetOrganizationsOK().WithPayload(organizationModels)
}

func (h *Handlers) getOrganization(params organizationOperations.GetOrganizationParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	var organization Organization

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	name := params.OrganizationName
	if err := h.store.Get(ctx, IdentityManagerFlags.OrgID, name, opts, &organization); err != nil {
		log.Errorf("store error when getting organization '%s': %+v", name, err)
		return organizationOperations.NewGetOrganizationNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("organization", name),
			})
	}

	organizationModel := organizationEntityToModel(&organization)

	return organizationOperations.NewGetOrganizationOK().WithPayload(organizationModel)
}

func (h *Handlers) addOrganization(params organizationOperations.AddOrganizationParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	organizationRequest := params.Body
	e := organizationModelToEntity(organizationRequest)

	e.Status = entitystore.StatusCREATING

	if _, err := h.store.Add(ctx, e); err != nil {
		if entitystore.IsUniqueViolation(err) {
			return organizationOperations.NewAddOrganizationConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("organization", e.Name),
			})
		}
		log.Errorf("store error when adding a new organization %s: %+v", e.Name, err)
		return organizationOperations.NewAddOrganizationDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("organization", e.Name),
		})
	}

	h.watcher.OnAction(ctx, e)

	return organizationOperations.NewAddOrganizationCreated().WithPayload(organizationEntityToModel(e))
}

func (h *Handlers) deleteOrganization(params organizationOperations.DeleteOrganizationParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	name := params.OrganizationName

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	var e Organization
	if err := h.store.Get(ctx, IdentityManagerFlags.OrgID, name, opts, &e); err != nil {
		log.Errorf("store error when getting organization: %+v", err)
		return organizationOperations.NewDeleteOrganizationNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("organization", e.Name),
			})
	}

	if e.Status == entitystore.StatusDELETING {
		log.Warnf("Attempting to delete organization  %s which already is in DELETING state: %+v", e.Name)
		return organizationOperations.NewDeleteOrganizationBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("Unable to delete organization %s: organization is already being deleted", e.Name)),
		})
	}

	e.Status = entitystore.StatusDELETING
	if _, err := h.store.Update(ctx, e.Revision, &e); err != nil {
		log.Errorf("store error when deleting a organization %s: %+v", e.Name, err)
		return organizationOperations.NewDeleteOrganizationDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("organization", e.Name),
		})
	}

	h.watcher.OnAction(ctx, &e)

	return organizationOperations.NewDeleteOrganizationOK().WithPayload(organizationEntityToModel(&e))
}

func (h *Handlers) updateOrganization(params organizationOperations.UpdateOrganizationParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	e := Organization{}
	if err := h.store.Get(ctx, IdentityManagerFlags.OrgID, params.OrganizationName, opts, &e); err != nil {
		log.Errorf("store error when getting organization: %+v", err)
		return organizationOperations.NewUpdateOrganizationNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("organization", e.Name),
			})
	}

	updateEntity := organizationModelToEntity(params.Body)
	updateEntity.CreatedTime = e.CreatedTime
	updateEntity.ID = e.ID
	updateEntity.Status = entitystore.StatusUPDATING

	if _, err := h.store.Update(ctx, e.Revision, updateEntity); err != nil {
		log.Errorf("store error when updating a organization %s: %+v", e.Name, err)
		return organizationOperations.NewUpdateOrganizationDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("organization", e.Name),
		})
	}

	h.watcher.OnAction(ctx, updateEntity)

	return organizationOperations.NewUpdateOrganizationOK().WithPayload(organizationEntityToModel(updateEntity))
}
