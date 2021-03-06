///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package store

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// UpdateFunctionOKCode is the HTTP code returned for type UpdateFunctionOK
const UpdateFunctionOKCode int = 200

/*UpdateFunctionOK Successful update

swagger:response updateFunctionOK
*/
type UpdateFunctionOK struct {

	/*
	  In: Body
	*/
	Payload *v1.Function `json:"body,omitempty"`
}

// NewUpdateFunctionOK creates UpdateFunctionOK with default headers values
func NewUpdateFunctionOK() *UpdateFunctionOK {

	return &UpdateFunctionOK{}
}

// WithPayload adds the payload to the update function o k response
func (o *UpdateFunctionOK) WithPayload(payload *v1.Function) *UpdateFunctionOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update function o k response
func (o *UpdateFunctionOK) SetPayload(payload *v1.Function) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateFunctionOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateFunctionBadRequestCode is the HTTP code returned for type UpdateFunctionBadRequest
const UpdateFunctionBadRequestCode int = 400

/*UpdateFunctionBadRequest Invalid input

swagger:response updateFunctionBadRequest
*/
type UpdateFunctionBadRequest struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateFunctionBadRequest creates UpdateFunctionBadRequest with default headers values
func NewUpdateFunctionBadRequest() *UpdateFunctionBadRequest {

	return &UpdateFunctionBadRequest{}
}

// WithPayload adds the payload to the update function bad request response
func (o *UpdateFunctionBadRequest) WithPayload(payload *v1.Error) *UpdateFunctionBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update function bad request response
func (o *UpdateFunctionBadRequest) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateFunctionBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateFunctionUnauthorizedCode is the HTTP code returned for type UpdateFunctionUnauthorized
const UpdateFunctionUnauthorizedCode int = 401

/*UpdateFunctionUnauthorized Unauthorized Request

swagger:response updateFunctionUnauthorized
*/
type UpdateFunctionUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateFunctionUnauthorized creates UpdateFunctionUnauthorized with default headers values
func NewUpdateFunctionUnauthorized() *UpdateFunctionUnauthorized {

	return &UpdateFunctionUnauthorized{}
}

// WithPayload adds the payload to the update function unauthorized response
func (o *UpdateFunctionUnauthorized) WithPayload(payload *v1.Error) *UpdateFunctionUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update function unauthorized response
func (o *UpdateFunctionUnauthorized) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateFunctionUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateFunctionForbiddenCode is the HTTP code returned for type UpdateFunctionForbidden
const UpdateFunctionForbiddenCode int = 403

/*UpdateFunctionForbidden access to this resource is forbidden

swagger:response updateFunctionForbidden
*/
type UpdateFunctionForbidden struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateFunctionForbidden creates UpdateFunctionForbidden with default headers values
func NewUpdateFunctionForbidden() *UpdateFunctionForbidden {

	return &UpdateFunctionForbidden{}
}

// WithPayload adds the payload to the update function forbidden response
func (o *UpdateFunctionForbidden) WithPayload(payload *v1.Error) *UpdateFunctionForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update function forbidden response
func (o *UpdateFunctionForbidden) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateFunctionForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateFunctionNotFoundCode is the HTTP code returned for type UpdateFunctionNotFound
const UpdateFunctionNotFoundCode int = 404

/*UpdateFunctionNotFound Function not found

swagger:response updateFunctionNotFound
*/
type UpdateFunctionNotFound struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateFunctionNotFound creates UpdateFunctionNotFound with default headers values
func NewUpdateFunctionNotFound() *UpdateFunctionNotFound {

	return &UpdateFunctionNotFound{}
}

// WithPayload adds the payload to the update function not found response
func (o *UpdateFunctionNotFound) WithPayload(payload *v1.Error) *UpdateFunctionNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update function not found response
func (o *UpdateFunctionNotFound) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateFunctionNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*UpdateFunctionDefault Unknown error

swagger:response updateFunctionDefault
*/
type UpdateFunctionDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateFunctionDefault creates UpdateFunctionDefault with default headers values
func NewUpdateFunctionDefault(code int) *UpdateFunctionDefault {
	if code <= 0 {
		code = 500
	}

	return &UpdateFunctionDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the update function default response
func (o *UpdateFunctionDefault) WithStatusCode(code int) *UpdateFunctionDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the update function default response
func (o *UpdateFunctionDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the update function default response
func (o *UpdateFunctionDefault) WithPayload(payload *v1.Error) *UpdateFunctionDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update function default response
func (o *UpdateFunctionDefault) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateFunctionDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
