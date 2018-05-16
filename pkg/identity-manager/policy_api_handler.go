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
	policyOperations "github.com/vmware/dispatch/pkg/identity-manager/gen/restapi/operations/policy"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

func policyModelToEntity(m *v1.Policy) *Policy {
	defer trace.Tracef("name '%s'", *m.Name)()

	e := Policy{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: IdentityManagerFlags.OrgID,
			Name:           *m.Name,
		},
	}
	for _, r := range m.Rules {
		rule := Rule{
			Subjects:  r.Subjects,
			Resources: r.Resources,
			Actions:   r.Actions,
		}
		e.Rules = append(e.Rules, rule)
	}
	return &e
}

func policyEntityToModel(e *Policy) *v1.Policy {
	defer trace.Tracef("name '%s'", e.Name)()
	m := v1.Policy{
		ID:           strfmt.UUID(e.ID),
		Name:         swag.String(e.Name),
		Kind:         utils.PolicyKind,
		Status:       v1.Status(e.Status),
		CreatedTime:  e.CreatedTime.Unix(),
		ModifiedTime: e.ModifiedTime.Unix(),
	}
	for _, r := range e.Rules {
		rule := v1.Rule{
			Subjects:  r.Subjects,
			Resources: r.Resources,
			Actions:   r.Actions,
		}
		m.Rules = append(m.Rules, &rule)
	}
	return &m
}

func (h *Handlers) getPolicies(params policyOperations.GetPoliciesParams, principal interface{}) middleware.Responder {

	defer trace.Trace("")()
	var policies []*Policy

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}
	err := h.store.List(IdentityManagerFlags.OrgID, opts, &policies)
	if err != nil {
		log.Errorf("store error when listing policies: %+v", err)
		return policyOperations.NewGetPoliciesInternalServerError().WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting policies"),
			})
	}
	var policyModels []*v1.Policy
	for _, policy := range policies {
		policyModels = append(policyModels, policyEntityToModel(policy))
	}
	return policyOperations.NewGetPoliciesOK().WithPayload(policyModels)
}

func (h *Handlers) getPolicy(params policyOperations.GetPolicyParams, principal interface{}) middleware.Responder {

	defer trace.Tracef("get policy name '%s'", params.PolicyName)()
	var policy Policy

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	name := params.PolicyName
	if err := h.store.Get(IdentityManagerFlags.OrgID, name, opts, &policy); err != nil {
		log.Errorf("store error when getting policy '%s': %+v", name, err)
		return policyOperations.NewGetPolicyNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("policy not found"),
			})
	}

	policyModel := policyEntityToModel(&policy)

	return policyOperations.NewGetPolicyOK().WithPayload(policyModel)
}

func (h *Handlers) addPolicy(params policyOperations.AddPolicyParams, principal interface{}) middleware.Responder {
	defer trace.Trace("")()
	policyRequest := params.Body
	e := policyModelToEntity(policyRequest)
	for _, rule := range e.Rules {
		// Do some basic validation although this must be handled at the goswagger server.
		if rule.Subjects == nil || rule.Actions == nil || rule.Resources == nil {
			return policyOperations.NewAddPolicyBadRequest().WithPayload(
				&v1.Error{
					Code:    http.StatusBadRequest,
					Message: swag.String("invalid rule definition, missing required fields"),
				})
		}
	}

	e.Status = entitystore.StatusCREATING

	if _, err := h.store.Add(e); err != nil {
		if entitystore.IsUniqueViolation(err) {
			return policyOperations.NewAddPolicyConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: swag.String("error creating policy: non-unique name"),
			})
		}
		log.Errorf("store error when adding a new policy %s: %+v", e.Name, err)
		return policyOperations.NewAddPolicyInternalServerError().WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing new policy"),
		})
	}

	h.watcher.OnAction(e)

	return policyOperations.NewAddPolicyCreated().WithPayload(policyEntityToModel(e))
}

func (h *Handlers) deletePolicy(params policyOperations.DeletePolicyParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.PolicyName)()
	name := params.PolicyName

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	var e Policy
	if err := h.store.Get(IdentityManagerFlags.OrgID, name, opts, &e); err != nil {
		log.Errorf("store error when getting policy: %+v", err)
		return policyOperations.NewDeletePolicyNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("policy not found"),
			})
	}

	if e.Status == entitystore.StatusDELETING {
		log.Warnf("Attempting to delete policy  %s which already is in DELETING state: %+v", e.Name)
		return policyOperations.NewDeletePolicyBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("Unable to delete policy %s: policy is already being deleted", e.Name)),
		})
	}

	e.Status = entitystore.StatusDELETING
	if _, err := h.store.Update(e.Revision, &e); err != nil {
		log.Errorf("store error when deleting a policy %s: %+v", e.Name, err)
		return policyOperations.NewDeletePolicyInternalServerError().WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting a policy"),
		})
	}

	h.watcher.OnAction(&e)

	return policyOperations.NewDeletePolicyOK().WithPayload(policyEntityToModel(&e))
}

func (h *Handlers) updatePolicy(params policyOperations.UpdatePolicyParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("updated policy '%s'", params.PolicyName)()

	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	e := Policy{}
	if err := h.store.Get(IdentityManagerFlags.OrgID, params.PolicyName, opts, &e); err != nil {
		log.Errorf("store error when getting policy: %+v", err)
		return policyOperations.NewUpdatePolicyNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("policy not found"),
			})
	}

	updateEntity := policyModelToEntity(params.Body)
	updateEntity.CreatedTime = e.CreatedTime
	updateEntity.ID = e.ID
	updateEntity.Status = entitystore.StatusUPDATING

	if _, err := h.store.Update(e.Revision, updateEntity); err != nil {
		log.Errorf("store error when updating a policy %s: %+v", e.Name, err)
		return policyOperations.NewUpdatePolicyInternalServerError().WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when updating a policy"),
		})
	}

	h.watcher.OnAction(updateEntity)

	return policyOperations.NewUpdatePolicyOK().WithPayload(policyEntityToModel(updateEntity))
}
