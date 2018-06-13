///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package policy

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// GetPoliciesOKCode is the HTTP code returned for type GetPoliciesOK
const GetPoliciesOKCode int = 200

/*GetPoliciesOK Successful operation

swagger:response getPoliciesOK
*/
type GetPoliciesOK struct {

	/*
	  In: Body
	*/
	Payload []*v1.Policy `json:"body,omitempty"`
}

// NewGetPoliciesOK creates GetPoliciesOK with default headers values
func NewGetPoliciesOK() *GetPoliciesOK {

	return &GetPoliciesOK{}
}

// WithPayload adds the payload to the get policies o k response
func (o *GetPoliciesOK) WithPayload(payload []*v1.Policy) *GetPoliciesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get policies o k response
func (o *GetPoliciesOK) SetPayload(payload []*v1.Policy) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPoliciesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		payload = make([]*v1.Policy, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}

}

// GetPoliciesUnauthorizedCode is the HTTP code returned for type GetPoliciesUnauthorized
const GetPoliciesUnauthorizedCode int = 401

/*GetPoliciesUnauthorized Unauthorized Request

swagger:response getPoliciesUnauthorized
*/
type GetPoliciesUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewGetPoliciesUnauthorized creates GetPoliciesUnauthorized with default headers values
func NewGetPoliciesUnauthorized() *GetPoliciesUnauthorized {

	return &GetPoliciesUnauthorized{}
}

// WithPayload adds the payload to the get policies unauthorized response
func (o *GetPoliciesUnauthorized) WithPayload(payload *v1.Error) *GetPoliciesUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get policies unauthorized response
func (o *GetPoliciesUnauthorized) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPoliciesUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetPoliciesForbiddenCode is the HTTP code returned for type GetPoliciesForbidden
const GetPoliciesForbiddenCode int = 403

/*GetPoliciesForbidden access to this resource is forbidden

swagger:response getPoliciesForbidden
*/
type GetPoliciesForbidden struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewGetPoliciesForbidden creates GetPoliciesForbidden with default headers values
func NewGetPoliciesForbidden() *GetPoliciesForbidden {

	return &GetPoliciesForbidden{}
}

// WithPayload adds the payload to the get policies forbidden response
func (o *GetPoliciesForbidden) WithPayload(payload *v1.Error) *GetPoliciesForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get policies forbidden response
func (o *GetPoliciesForbidden) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPoliciesForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetPoliciesDefault Unexpected Error

swagger:response getPoliciesDefault
*/
type GetPoliciesDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewGetPoliciesDefault creates GetPoliciesDefault with default headers values
func NewGetPoliciesDefault(code int) *GetPoliciesDefault {
	if code <= 0 {
		code = 500
	}

	return &GetPoliciesDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get policies default response
func (o *GetPoliciesDefault) WithStatusCode(code int) *GetPoliciesDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get policies default response
func (o *GetPoliciesDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get policies default response
func (o *GetPoliciesDefault) WithPayload(payload *v1.Error) *GetPoliciesDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get policies default response
func (o *GetPoliciesDefault) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPoliciesDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
