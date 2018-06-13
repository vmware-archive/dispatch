///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package subscriptions

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// DeleteSubscriptionOKCode is the HTTP code returned for type DeleteSubscriptionOK
const DeleteSubscriptionOKCode int = 200

/*DeleteSubscriptionOK successful operation

swagger:response deleteSubscriptionOK
*/
type DeleteSubscriptionOK struct {

	/*
	  In: Body
	*/
	Payload *v1.Subscription `json:"body,omitempty"`
}

// NewDeleteSubscriptionOK creates DeleteSubscriptionOK with default headers values
func NewDeleteSubscriptionOK() *DeleteSubscriptionOK {

	return &DeleteSubscriptionOK{}
}

// WithPayload adds the payload to the delete subscription o k response
func (o *DeleteSubscriptionOK) WithPayload(payload *v1.Subscription) *DeleteSubscriptionOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete subscription o k response
func (o *DeleteSubscriptionOK) SetPayload(payload *v1.Subscription) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteSubscriptionOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteSubscriptionBadRequestCode is the HTTP code returned for type DeleteSubscriptionBadRequest
const DeleteSubscriptionBadRequestCode int = 400

/*DeleteSubscriptionBadRequest Invalid ID supplied

swagger:response deleteSubscriptionBadRequest
*/
type DeleteSubscriptionBadRequest struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewDeleteSubscriptionBadRequest creates DeleteSubscriptionBadRequest with default headers values
func NewDeleteSubscriptionBadRequest() *DeleteSubscriptionBadRequest {

	return &DeleteSubscriptionBadRequest{}
}

// WithPayload adds the payload to the delete subscription bad request response
func (o *DeleteSubscriptionBadRequest) WithPayload(payload *v1.Error) *DeleteSubscriptionBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete subscription bad request response
func (o *DeleteSubscriptionBadRequest) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteSubscriptionBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteSubscriptionUnauthorizedCode is the HTTP code returned for type DeleteSubscriptionUnauthorized
const DeleteSubscriptionUnauthorizedCode int = 401

/*DeleteSubscriptionUnauthorized Unauthorized Request

swagger:response deleteSubscriptionUnauthorized
*/
type DeleteSubscriptionUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewDeleteSubscriptionUnauthorized creates DeleteSubscriptionUnauthorized with default headers values
func NewDeleteSubscriptionUnauthorized() *DeleteSubscriptionUnauthorized {

	return &DeleteSubscriptionUnauthorized{}
}

// WithPayload adds the payload to the delete subscription unauthorized response
func (o *DeleteSubscriptionUnauthorized) WithPayload(payload *v1.Error) *DeleteSubscriptionUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete subscription unauthorized response
func (o *DeleteSubscriptionUnauthorized) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteSubscriptionUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteSubscriptionForbiddenCode is the HTTP code returned for type DeleteSubscriptionForbidden
const DeleteSubscriptionForbiddenCode int = 403

/*DeleteSubscriptionForbidden access to this resource is forbidden

swagger:response deleteSubscriptionForbidden
*/
type DeleteSubscriptionForbidden struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewDeleteSubscriptionForbidden creates DeleteSubscriptionForbidden with default headers values
func NewDeleteSubscriptionForbidden() *DeleteSubscriptionForbidden {

	return &DeleteSubscriptionForbidden{}
}

// WithPayload adds the payload to the delete subscription forbidden response
func (o *DeleteSubscriptionForbidden) WithPayload(payload *v1.Error) *DeleteSubscriptionForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete subscription forbidden response
func (o *DeleteSubscriptionForbidden) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteSubscriptionForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteSubscriptionNotFoundCode is the HTTP code returned for type DeleteSubscriptionNotFound
const DeleteSubscriptionNotFoundCode int = 404

/*DeleteSubscriptionNotFound Subscription not found

swagger:response deleteSubscriptionNotFound
*/
type DeleteSubscriptionNotFound struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewDeleteSubscriptionNotFound creates DeleteSubscriptionNotFound with default headers values
func NewDeleteSubscriptionNotFound() *DeleteSubscriptionNotFound {

	return &DeleteSubscriptionNotFound{}
}

// WithPayload adds the payload to the delete subscription not found response
func (o *DeleteSubscriptionNotFound) WithPayload(payload *v1.Error) *DeleteSubscriptionNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete subscription not found response
func (o *DeleteSubscriptionNotFound) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteSubscriptionNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*DeleteSubscriptionDefault Generic error response

swagger:response deleteSubscriptionDefault
*/
type DeleteSubscriptionDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewDeleteSubscriptionDefault creates DeleteSubscriptionDefault with default headers values
func NewDeleteSubscriptionDefault(code int) *DeleteSubscriptionDefault {
	if code <= 0 {
		code = 500
	}

	return &DeleteSubscriptionDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the delete subscription default response
func (o *DeleteSubscriptionDefault) WithStatusCode(code int) *DeleteSubscriptionDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the delete subscription default response
func (o *DeleteSubscriptionDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the delete subscription default response
func (o *DeleteSubscriptionDefault) WithPayload(payload *v1.Error) *DeleteSubscriptionDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete subscription default response
func (o *DeleteSubscriptionDefault) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteSubscriptionDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
