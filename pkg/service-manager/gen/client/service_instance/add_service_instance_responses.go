///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package service_instance

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware/dispatch/pkg/service-manager/gen/models"
)

// AddServiceInstanceReader is a Reader for the AddServiceInstance structure.
type AddServiceInstanceReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *AddServiceInstanceReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 201:
		result := NewAddServiceInstanceCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewAddServiceInstanceBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 409:
		result := NewAddServiceInstanceConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewAddServiceInstanceDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewAddServiceInstanceCreated creates a AddServiceInstanceCreated with default headers values
func NewAddServiceInstanceCreated() *AddServiceInstanceCreated {
	return &AddServiceInstanceCreated{}
}

/*AddServiceInstanceCreated handles this case with default header values.

created
*/
type AddServiceInstanceCreated struct {
	Payload *models.ServiceInstance
}

func (o *AddServiceInstanceCreated) Error() string {
	return fmt.Sprintf("[POST /serviceinstance][%d] addServiceInstanceCreated  %+v", 201, o.Payload)
}

func (o *AddServiceInstanceCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ServiceInstance)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddServiceInstanceBadRequest creates a AddServiceInstanceBadRequest with default headers values
func NewAddServiceInstanceBadRequest() *AddServiceInstanceBadRequest {
	return &AddServiceInstanceBadRequest{}
}

/*AddServiceInstanceBadRequest handles this case with default header values.

Invalid input
*/
type AddServiceInstanceBadRequest struct {
	Payload *models.Error
}

func (o *AddServiceInstanceBadRequest) Error() string {
	return fmt.Sprintf("[POST /serviceinstance][%d] addServiceInstanceBadRequest  %+v", 400, o.Payload)
}

func (o *AddServiceInstanceBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddServiceInstanceConflict creates a AddServiceInstanceConflict with default headers values
func NewAddServiceInstanceConflict() *AddServiceInstanceConflict {
	return &AddServiceInstanceConflict{}
}

/*AddServiceInstanceConflict handles this case with default header values.

Already Exists
*/
type AddServiceInstanceConflict struct {
	Payload *models.Error
}

func (o *AddServiceInstanceConflict) Error() string {
	return fmt.Sprintf("[POST /serviceinstance][%d] addServiceInstanceConflict  %+v", 409, o.Payload)
}

func (o *AddServiceInstanceConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddServiceInstanceDefault creates a AddServiceInstanceDefault with default headers values
func NewAddServiceInstanceDefault(code int) *AddServiceInstanceDefault {
	return &AddServiceInstanceDefault{
		_statusCode: code,
	}
}

/*AddServiceInstanceDefault handles this case with default header values.

Generic error response
*/
type AddServiceInstanceDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the add service instance default response
func (o *AddServiceInstanceDefault) Code() int {
	return o._statusCode
}

func (o *AddServiceInstanceDefault) Error() string {
	return fmt.Sprintf("[POST /serviceinstance][%d] addServiceInstance default  %+v", o._statusCode, o.Payload)
}

func (o *AddServiceInstanceDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
