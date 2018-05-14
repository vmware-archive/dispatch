///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package drivers

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// UpdateDriverTypeOKCode is the HTTP code returned for type UpdateDriverTypeOK
const UpdateDriverTypeOKCode int = 200

/*UpdateDriverTypeOK Successful operation

swagger:response updateDriverTypeOK
*/
type UpdateDriverTypeOK struct {

	/*
	  In: Body
	*/
	Payload *v1.EventDriverType `json:"body,omitempty"`
}

// NewUpdateDriverTypeOK creates UpdateDriverTypeOK with default headers values
func NewUpdateDriverTypeOK() *UpdateDriverTypeOK {

	return &UpdateDriverTypeOK{}
}

// WithPayload adds the payload to the update driver type o k response
func (o *UpdateDriverTypeOK) WithPayload(payload *v1.EventDriverType) *UpdateDriverTypeOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update driver type o k response
func (o *UpdateDriverTypeOK) SetPayload(payload *v1.EventDriverType) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateDriverTypeOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateDriverTypeBadRequestCode is the HTTP code returned for type UpdateDriverTypeBadRequest
const UpdateDriverTypeBadRequestCode int = 400

/*UpdateDriverTypeBadRequest Invalid Name supplied

swagger:response updateDriverTypeBadRequest
*/
type UpdateDriverTypeBadRequest struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateDriverTypeBadRequest creates UpdateDriverTypeBadRequest with default headers values
func NewUpdateDriverTypeBadRequest() *UpdateDriverTypeBadRequest {

	return &UpdateDriverTypeBadRequest{}
}

// WithPayload adds the payload to the update driver type bad request response
func (o *UpdateDriverTypeBadRequest) WithPayload(payload *v1.Error) *UpdateDriverTypeBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update driver type bad request response
func (o *UpdateDriverTypeBadRequest) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateDriverTypeBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateDriverTypeNotFoundCode is the HTTP code returned for type UpdateDriverTypeNotFound
const UpdateDriverTypeNotFoundCode int = 404

/*UpdateDriverTypeNotFound DriverType not found

swagger:response updateDriverTypeNotFound
*/
type UpdateDriverTypeNotFound struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateDriverTypeNotFound creates UpdateDriverTypeNotFound with default headers values
func NewUpdateDriverTypeNotFound() *UpdateDriverTypeNotFound {

	return &UpdateDriverTypeNotFound{}
}

// WithPayload adds the payload to the update driver type not found response
func (o *UpdateDriverTypeNotFound) WithPayload(payload *v1.Error) *UpdateDriverTypeNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update driver type not found response
func (o *UpdateDriverTypeNotFound) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateDriverTypeNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateDriverTypeInternalServerErrorCode is the HTTP code returned for type UpdateDriverTypeInternalServerError
const UpdateDriverTypeInternalServerErrorCode int = 500

/*UpdateDriverTypeInternalServerError Internal server error

swagger:response updateDriverTypeInternalServerError
*/
type UpdateDriverTypeInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateDriverTypeInternalServerError creates UpdateDriverTypeInternalServerError with default headers values
func NewUpdateDriverTypeInternalServerError() *UpdateDriverTypeInternalServerError {

	return &UpdateDriverTypeInternalServerError{}
}

// WithPayload adds the payload to the update driver type internal server error response
func (o *UpdateDriverTypeInternalServerError) WithPayload(payload *v1.Error) *UpdateDriverTypeInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update driver type internal server error response
func (o *UpdateDriverTypeInternalServerError) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateDriverTypeInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*UpdateDriverTypeDefault Unknown error

swagger:response updateDriverTypeDefault
*/
type UpdateDriverTypeDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewUpdateDriverTypeDefault creates UpdateDriverTypeDefault with default headers values
func NewUpdateDriverTypeDefault(code int) *UpdateDriverTypeDefault {
	if code <= 0 {
		code = 500
	}

	return &UpdateDriverTypeDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the update driver type default response
func (o *UpdateDriverTypeDefault) WithStatusCode(code int) *UpdateDriverTypeDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the update driver type default response
func (o *UpdateDriverTypeDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the update driver type default response
func (o *UpdateDriverTypeDefault) WithPayload(payload *v1.Error) *UpdateDriverTypeDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update driver type default response
func (o *UpdateDriverTypeDefault) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateDriverTypeDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
