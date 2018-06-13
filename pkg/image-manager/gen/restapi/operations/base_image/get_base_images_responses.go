///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package base_image

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// GetBaseImagesOKCode is the HTTP code returned for type GetBaseImagesOK
const GetBaseImagesOKCode int = 200

/*GetBaseImagesOK successful operation

swagger:response getBaseImagesOK
*/
type GetBaseImagesOK struct {

	/*
	  In: Body
	*/
	Payload []*v1.BaseImage `json:"body,omitempty"`
}

// NewGetBaseImagesOK creates GetBaseImagesOK with default headers values
func NewGetBaseImagesOK() *GetBaseImagesOK {

	return &GetBaseImagesOK{}
}

// WithPayload adds the payload to the get base images o k response
func (o *GetBaseImagesOK) WithPayload(payload []*v1.BaseImage) *GetBaseImagesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get base images o k response
func (o *GetBaseImagesOK) SetPayload(payload []*v1.BaseImage) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetBaseImagesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		payload = make([]*v1.BaseImage, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}

}

// GetBaseImagesUnauthorizedCode is the HTTP code returned for type GetBaseImagesUnauthorized
const GetBaseImagesUnauthorizedCode int = 401

/*GetBaseImagesUnauthorized Unauthorized Request

swagger:response getBaseImagesUnauthorized
*/
type GetBaseImagesUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewGetBaseImagesUnauthorized creates GetBaseImagesUnauthorized with default headers values
func NewGetBaseImagesUnauthorized() *GetBaseImagesUnauthorized {

	return &GetBaseImagesUnauthorized{}
}

// WithPayload adds the payload to the get base images unauthorized response
func (o *GetBaseImagesUnauthorized) WithPayload(payload *v1.Error) *GetBaseImagesUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get base images unauthorized response
func (o *GetBaseImagesUnauthorized) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetBaseImagesUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetBaseImagesForbiddenCode is the HTTP code returned for type GetBaseImagesForbidden
const GetBaseImagesForbiddenCode int = 403

/*GetBaseImagesForbidden access to this resource is forbidden

swagger:response getBaseImagesForbidden
*/
type GetBaseImagesForbidden struct {

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewGetBaseImagesForbidden creates GetBaseImagesForbidden with default headers values
func NewGetBaseImagesForbidden() *GetBaseImagesForbidden {

	return &GetBaseImagesForbidden{}
}

// WithPayload adds the payload to the get base images forbidden response
func (o *GetBaseImagesForbidden) WithPayload(payload *v1.Error) *GetBaseImagesForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get base images forbidden response
func (o *GetBaseImagesForbidden) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetBaseImagesForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetBaseImagesDefault Generic error response

swagger:response getBaseImagesDefault
*/
type GetBaseImagesDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *v1.Error `json:"body,omitempty"`
}

// NewGetBaseImagesDefault creates GetBaseImagesDefault with default headers values
func NewGetBaseImagesDefault(code int) *GetBaseImagesDefault {
	if code <= 0 {
		code = 500
	}

	return &GetBaseImagesDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get base images default response
func (o *GetBaseImagesDefault) WithStatusCode(code int) *GetBaseImagesDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get base images default response
func (o *GetBaseImagesDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get base images default response
func (o *GetBaseImagesDefault) WithPayload(payload *v1.Error) *GetBaseImagesDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get base images default response
func (o *GetBaseImagesDefault) SetPayload(payload *v1.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetBaseImagesDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
