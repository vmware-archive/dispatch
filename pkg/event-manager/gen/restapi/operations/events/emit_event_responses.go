///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package events

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

// EmitEventOKCode is the HTTP code returned for type EmitEventOK
const EmitEventOKCode int = 200

/*EmitEventOK Event emitted

swagger:response emitEventOK
*/
type EmitEventOK struct {

	/*
	  In: Body
	*/
	Payload *models.Emission `json:"body,omitempty"`
}

// NewEmitEventOK creates EmitEventOK with default headers values
func NewEmitEventOK() *EmitEventOK {
	return &EmitEventOK{}
}

// WithPayload adds the payload to the emit event o k response
func (o *EmitEventOK) WithPayload(payload *models.Emission) *EmitEventOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the emit event o k response
func (o *EmitEventOK) SetPayload(payload *models.Emission) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *EmitEventOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// EmitEventBadRequestCode is the HTTP code returned for type EmitEventBadRequest
const EmitEventBadRequestCode int = 400

/*EmitEventBadRequest Invalid input

swagger:response emitEventBadRequest
*/
type EmitEventBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewEmitEventBadRequest creates EmitEventBadRequest with default headers values
func NewEmitEventBadRequest() *EmitEventBadRequest {
	return &EmitEventBadRequest{}
}

// WithPayload adds the payload to the emit event bad request response
func (o *EmitEventBadRequest) WithPayload(payload *models.Error) *EmitEventBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the emit event bad request response
func (o *EmitEventBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *EmitEventBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// EmitEventUnauthorizedCode is the HTTP code returned for type EmitEventUnauthorized
const EmitEventUnauthorizedCode int = 401

/*EmitEventUnauthorized Unauthorized Request

swagger:response emitEventUnauthorized
*/
type EmitEventUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewEmitEventUnauthorized creates EmitEventUnauthorized with default headers values
func NewEmitEventUnauthorized() *EmitEventUnauthorized {
	return &EmitEventUnauthorized{}
}

// WithPayload adds the payload to the emit event unauthorized response
func (o *EmitEventUnauthorized) WithPayload(payload *models.Error) *EmitEventUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the emit event unauthorized response
func (o *EmitEventUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *EmitEventUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// EmitEventInternalServerErrorCode is the HTTP code returned for type EmitEventInternalServerError
const EmitEventInternalServerErrorCode int = 500

/*EmitEventInternalServerError Internal server error

swagger:response emitEventInternalServerError
*/
type EmitEventInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewEmitEventInternalServerError creates EmitEventInternalServerError with default headers values
func NewEmitEventInternalServerError() *EmitEventInternalServerError {
	return &EmitEventInternalServerError{}
}

// WithPayload adds the payload to the emit event internal server error response
func (o *EmitEventInternalServerError) WithPayload(payload *models.Error) *EmitEventInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the emit event internal server error response
func (o *EmitEventInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *EmitEventInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*EmitEventDefault Unknown error

swagger:response emitEventDefault
*/
type EmitEventDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewEmitEventDefault creates EmitEventDefault with default headers values
func NewEmitEventDefault(code int) *EmitEventDefault {
	if code <= 0 {
		code = 500
	}

	return &EmitEventDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the emit event default response
func (o *EmitEventDefault) WithStatusCode(code int) *EmitEventDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the emit event default response
func (o *EmitEventDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the emit event default response
func (o *EmitEventDefault) WithPayload(payload *models.Error) *EmitEventDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the emit event default response
func (o *EmitEventDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *EmitEventDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
