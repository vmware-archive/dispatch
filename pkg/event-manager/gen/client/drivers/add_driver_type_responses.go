///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package drivers

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// AddDriverTypeReader is a Reader for the AddDriverType structure.
type AddDriverTypeReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *AddDriverTypeReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 201:
		result := NewAddDriverTypeCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewAddDriverTypeBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 401:
		result := NewAddDriverTypeUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 403:
		result := NewAddDriverTypeForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 409:
		result := NewAddDriverTypeConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewAddDriverTypeDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewAddDriverTypeCreated creates a AddDriverTypeCreated with default headers values
func NewAddDriverTypeCreated() *AddDriverTypeCreated {
	return &AddDriverTypeCreated{}
}

/*AddDriverTypeCreated handles this case with default header values.

Driver Type created
*/
type AddDriverTypeCreated struct {
	Payload *v1.EventDriverType
}

func (o *AddDriverTypeCreated) Error() string {
	return fmt.Sprintf("[POST /drivertypes][%d] addDriverTypeCreated  %+v", 201, o.Payload)
}

func (o *AddDriverTypeCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.EventDriverType)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddDriverTypeBadRequest creates a AddDriverTypeBadRequest with default headers values
func NewAddDriverTypeBadRequest() *AddDriverTypeBadRequest {
	return &AddDriverTypeBadRequest{}
}

/*AddDriverTypeBadRequest handles this case with default header values.

Invalid input
*/
type AddDriverTypeBadRequest struct {
	Payload *v1.Error
}

func (o *AddDriverTypeBadRequest) Error() string {
	return fmt.Sprintf("[POST /drivertypes][%d] addDriverTypeBadRequest  %+v", 400, o.Payload)
}

func (o *AddDriverTypeBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddDriverTypeUnauthorized creates a AddDriverTypeUnauthorized with default headers values
func NewAddDriverTypeUnauthorized() *AddDriverTypeUnauthorized {
	return &AddDriverTypeUnauthorized{}
}

/*AddDriverTypeUnauthorized handles this case with default header values.

Unauthorized Request
*/
type AddDriverTypeUnauthorized struct {
	Payload *v1.Error
}

func (o *AddDriverTypeUnauthorized) Error() string {
	return fmt.Sprintf("[POST /drivertypes][%d] addDriverTypeUnauthorized  %+v", 401, o.Payload)
}

func (o *AddDriverTypeUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddDriverTypeForbidden creates a AddDriverTypeForbidden with default headers values
func NewAddDriverTypeForbidden() *AddDriverTypeForbidden {
	return &AddDriverTypeForbidden{}
}

/*AddDriverTypeForbidden handles this case with default header values.

access to this resource is forbidden
*/
type AddDriverTypeForbidden struct {
	Payload *v1.Error
}

func (o *AddDriverTypeForbidden) Error() string {
	return fmt.Sprintf("[POST /drivertypes][%d] addDriverTypeForbidden  %+v", 403, o.Payload)
}

func (o *AddDriverTypeForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddDriverTypeConflict creates a AddDriverTypeConflict with default headers values
func NewAddDriverTypeConflict() *AddDriverTypeConflict {
	return &AddDriverTypeConflict{}
}

/*AddDriverTypeConflict handles this case with default header values.

Already Exists
*/
type AddDriverTypeConflict struct {
	Payload *v1.Error
}

func (o *AddDriverTypeConflict) Error() string {
	return fmt.Sprintf("[POST /drivertypes][%d] addDriverTypeConflict  %+v", 409, o.Payload)
}

func (o *AddDriverTypeConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddDriverTypeDefault creates a AddDriverTypeDefault with default headers values
func NewAddDriverTypeDefault(code int) *AddDriverTypeDefault {
	return &AddDriverTypeDefault{
		_statusCode: code,
	}
}

/*AddDriverTypeDefault handles this case with default header values.

Unknown error
*/
type AddDriverTypeDefault struct {
	_statusCode int

	Payload *v1.Error
}

// Code gets the status code for the add driver type default response
func (o *AddDriverTypeDefault) Code() int {
	return o._statusCode
}

func (o *AddDriverTypeDefault) Error() string {
	return fmt.Sprintf("[POST /drivertypes][%d] addDriverType default  %+v", o._statusCode, o.Payload)
}

func (o *AddDriverTypeDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
