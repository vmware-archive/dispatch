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

	models "github.com/vmware/dispatch/pkg/function-manager/gen/models"
)

// GetFunctionsOKCode is the HTTP code returned for type GetFunctionsOK
const GetFunctionsOKCode int = 200

/*GetFunctionsOK Successful operation

swagger:response getFunctionsOK
*/
type GetFunctionsOK struct {

	/*
	  In: Body
	*/
	Payload []*models.Function `json:"body,omitempty"`
}

// NewGetFunctionsOK creates GetFunctionsOK with default headers values
func NewGetFunctionsOK() *GetFunctionsOK {

	return &GetFunctionsOK{}
}

// WithPayload adds the payload to the get functions o k response
func (o *GetFunctionsOK) WithPayload(payload []*models.Function) *GetFunctionsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get functions o k response
func (o *GetFunctionsOK) SetPayload(payload []*models.Function) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetFunctionsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		payload = make([]*models.Function, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}

}

// GetFunctionsBadRequestCode is the HTTP code returned for type GetFunctionsBadRequest
const GetFunctionsBadRequestCode int = 400

/*GetFunctionsBadRequest Invalid input

swagger:response getFunctionsBadRequest
*/
type GetFunctionsBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetFunctionsBadRequest creates GetFunctionsBadRequest with default headers values
func NewGetFunctionsBadRequest() *GetFunctionsBadRequest {

	return &GetFunctionsBadRequest{}
}

// WithPayload adds the payload to the get functions bad request response
func (o *GetFunctionsBadRequest) WithPayload(payload *models.Error) *GetFunctionsBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get functions bad request response
func (o *GetFunctionsBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetFunctionsBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetFunctionsInternalServerErrorCode is the HTTP code returned for type GetFunctionsInternalServerError
const GetFunctionsInternalServerErrorCode int = 500

/*GetFunctionsInternalServerError Internal error

swagger:response getFunctionsInternalServerError
*/
type GetFunctionsInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetFunctionsInternalServerError creates GetFunctionsInternalServerError with default headers values
func NewGetFunctionsInternalServerError() *GetFunctionsInternalServerError {

	return &GetFunctionsInternalServerError{}
}

// WithPayload adds the payload to the get functions internal server error response
func (o *GetFunctionsInternalServerError) WithPayload(payload *models.Error) *GetFunctionsInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get functions internal server error response
func (o *GetFunctionsInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetFunctionsInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetFunctionsDefault Custom error

swagger:response getFunctionsDefault
*/
type GetFunctionsDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetFunctionsDefault creates GetFunctionsDefault with default headers values
func NewGetFunctionsDefault(code int) *GetFunctionsDefault {
	if code <= 0 {
		code = 500
	}

	return &GetFunctionsDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get functions default response
func (o *GetFunctionsDefault) WithStatusCode(code int) *GetFunctionsDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get functions default response
func (o *GetFunctionsDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get functions default response
func (o *GetFunctionsDefault) WithPayload(payload *models.Error) *GetFunctionsDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get functions default response
func (o *GetFunctionsDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetFunctionsDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
