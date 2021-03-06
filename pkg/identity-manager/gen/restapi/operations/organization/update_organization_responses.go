///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package organization

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// UpdateOrganizationOKCode is the HTTP code returned for type UpdateOrganizationOK
const UpdateOrganizationOKCode int = 200

/*UpdateOrganizationOK Successful update

swagger:response updateOrganizationOK
*/
type UpdateOrganizationOK struct {

	/*
	  In: Body
	*/
	Payload *v1.Organization `json:"body,omitempty"`
}

// NewUpdateOrganizationOK creates UpdateOrganizationOK with default headers values
func NewUpdateOrganizationOK() *UpdateOrganizationOK {

	return &UpdateOrganizationOK{}
}

// WithPayload adds the payload to the update organization o k response
func (o *UpdateOrganizationOK) WithPayload(payload *v1.Organization) *UpdateOrganizationOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update organization o k response
func (o *UpdateOrganizationOK) SetPayload(payload *v1.Organization) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateOrganizationOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateOrganizationBadRequestCode is the HTTP code returned for type UpdateOrganizationBadRequest
const UpdateOrganizationBadRequestCode int = 400

/*UpdateOrganizationBadRequest Invalid input

swagger:response updateOrganizationBadRequest
*/
type UpdateOrganizationBadRequest struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateOrganizationBadRequest creates UpdateOrganizationBadRequest with default headers values
func NewUpdateOrganizationBadRequest() *UpdateOrganizationBadRequest {

	return &UpdateOrganizationBadRequest{}
}

// WithPayload adds the payload to the update organization bad request response
func (o *UpdateOrganizationBadRequest) WithPayload(payload *v1.Error) *UpdateOrganizationBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update organization bad request response
func (o *UpdateOrganizationBadRequest) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateOrganizationBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateOrganizationUnauthorizedCode is the HTTP code returned for type UpdateOrganizationUnauthorized
const UpdateOrganizationUnauthorizedCode int = 401

/*UpdateOrganizationUnauthorized Unauthorized Request

swagger:response updateOrganizationUnauthorized
*/
type UpdateOrganizationUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateOrganizationUnauthorized creates UpdateOrganizationUnauthorized with default headers values
func NewUpdateOrganizationUnauthorized() *UpdateOrganizationUnauthorized {

	return &UpdateOrganizationUnauthorized{}
}

// WithPayload adds the payload to the update organization unauthorized response
func (o *UpdateOrganizationUnauthorized) WithPayload(payload *v1.Error) *UpdateOrganizationUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update organization unauthorized response
func (o *UpdateOrganizationUnauthorized) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateOrganizationUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateOrganizationForbiddenCode is the HTTP code returned for type UpdateOrganizationForbidden
const UpdateOrganizationForbiddenCode int = 403

/*UpdateOrganizationForbidden access to this resource is forbidden

swagger:response updateOrganizationForbidden
*/
type UpdateOrganizationForbidden struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateOrganizationForbidden creates UpdateOrganizationForbidden with default headers values
func NewUpdateOrganizationForbidden() *UpdateOrganizationForbidden {

	return &UpdateOrganizationForbidden{}
}

// WithPayload adds the payload to the update organization forbidden response
func (o *UpdateOrganizationForbidden) WithPayload(payload *v1.Error) *UpdateOrganizationForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update organization forbidden response
func (o *UpdateOrganizationForbidden) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateOrganizationForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateOrganizationNotFoundCode is the HTTP code returned for type UpdateOrganizationNotFound
const UpdateOrganizationNotFoundCode int = 404

/*UpdateOrganizationNotFound Organization not found

swagger:response updateOrganizationNotFound
*/
type UpdateOrganizationNotFound struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateOrganizationNotFound creates UpdateOrganizationNotFound with default headers values
func NewUpdateOrganizationNotFound() *UpdateOrganizationNotFound {

	return &UpdateOrganizationNotFound{}
}

// WithPayload adds the payload to the update organization not found response
func (o *UpdateOrganizationNotFound) WithPayload(payload *v1.Error) *UpdateOrganizationNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update organization not found response
func (o *UpdateOrganizationNotFound) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateOrganizationNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*UpdateOrganizationDefault Unknown error

swagger:response updateOrganizationDefault
*/
type UpdateOrganizationDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateOrganizationDefault creates UpdateOrganizationDefault with default headers values
func NewUpdateOrganizationDefault(code int) *UpdateOrganizationDefault {
	if code <= 0 {
		code = 500
	}

	return &UpdateOrganizationDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the update organization default response
func (o *UpdateOrganizationDefault) WithStatusCode(code int) *UpdateOrganizationDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the update organization default response
func (o *UpdateOrganizationDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the update organization default response
func (o *UpdateOrganizationDefault) WithPayload(payload *v1.Error) *UpdateOrganizationDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update organization default response
func (o *UpdateOrganizationDefault) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateOrganizationDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
