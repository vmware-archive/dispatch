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

	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

// GetSubscriptionsOKCode is the HTTP code returned for type GetSubscriptionsOK
const GetSubscriptionsOKCode int = 200

/*GetSubscriptionsOK Successful operation

swagger:response getSubscriptionsOK
*/
type GetSubscriptionsOK struct {

	/*
	  In: Body
	*/
	Payload []*models.Subscription `json:"body,omitempty"`
}

// NewGetSubscriptionsOK creates GetSubscriptionsOK with default headers values
func NewGetSubscriptionsOK() *GetSubscriptionsOK {

	return &GetSubscriptionsOK{}
}

// WithPayload adds the payload to the get subscriptions o k response
func (o *GetSubscriptionsOK) WithPayload(payload []*models.Subscription) *GetSubscriptionsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get subscriptions o k response
func (o *GetSubscriptionsOK) SetPayload(payload []*models.Subscription) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSubscriptionsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		payload = make([]*models.Subscription, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}

}

// GetSubscriptionsBadRequestCode is the HTTP code returned for type GetSubscriptionsBadRequest
const GetSubscriptionsBadRequestCode int = 400

/*GetSubscriptionsBadRequest Bad Request

swagger:response getSubscriptionsBadRequest
*/
type GetSubscriptionsBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetSubscriptionsBadRequest creates GetSubscriptionsBadRequest with default headers values
func NewGetSubscriptionsBadRequest() *GetSubscriptionsBadRequest {

	return &GetSubscriptionsBadRequest{}
}

// WithPayload adds the payload to the get subscriptions bad request response
func (o *GetSubscriptionsBadRequest) WithPayload(payload *models.Error) *GetSubscriptionsBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get subscriptions bad request response
func (o *GetSubscriptionsBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSubscriptionsBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetSubscriptionsInternalServerErrorCode is the HTTP code returned for type GetSubscriptionsInternalServerError
const GetSubscriptionsInternalServerErrorCode int = 500

/*GetSubscriptionsInternalServerError Internal server error

swagger:response getSubscriptionsInternalServerError
*/
type GetSubscriptionsInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetSubscriptionsInternalServerError creates GetSubscriptionsInternalServerError with default headers values
func NewGetSubscriptionsInternalServerError() *GetSubscriptionsInternalServerError {

	return &GetSubscriptionsInternalServerError{}
}

// WithPayload adds the payload to the get subscriptions internal server error response
func (o *GetSubscriptionsInternalServerError) WithPayload(payload *models.Error) *GetSubscriptionsInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get subscriptions internal server error response
func (o *GetSubscriptionsInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSubscriptionsInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetSubscriptionsDefault Unknown error

swagger:response getSubscriptionsDefault
*/
type GetSubscriptionsDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetSubscriptionsDefault creates GetSubscriptionsDefault with default headers values
func NewGetSubscriptionsDefault(code int) *GetSubscriptionsDefault {
	if code <= 0 {
		code = 500
	}

	return &GetSubscriptionsDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get subscriptions default response
func (o *GetSubscriptionsDefault) WithStatusCode(code int) *GetSubscriptionsDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get subscriptions default response
func (o *GetSubscriptionsDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get subscriptions default response
func (o *GetSubscriptionsDefault) WithPayload(payload *models.Error) *GetSubscriptionsDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get subscriptions default response
func (o *GetSubscriptionsDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSubscriptionsDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
