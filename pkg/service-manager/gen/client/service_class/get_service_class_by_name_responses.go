///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package service_class

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware/dispatch/pkg/service-manager/gen/models"
)

// GetServiceClassByNameReader is a Reader for the GetServiceClassByName structure.
type GetServiceClassByNameReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetServiceClassByNameReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetServiceClassByNameOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewGetServiceClassByNameBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewGetServiceClassByNameNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewGetServiceClassByNameDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetServiceClassByNameOK creates a GetServiceClassByNameOK with default headers values
func NewGetServiceClassByNameOK() *GetServiceClassByNameOK {
	return &GetServiceClassByNameOK{}
}

/*GetServiceClassByNameOK handles this case with default header values.

successful operation
*/
type GetServiceClassByNameOK struct {
	Payload *models.ServiceClass
}

func (o *GetServiceClassByNameOK) Error() string {
	return fmt.Sprintf("[GET /serviceclass/{serviceClassName}][%d] getServiceClassByNameOK  %+v", 200, o.Payload)
}

func (o *GetServiceClassByNameOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ServiceClass)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetServiceClassByNameBadRequest creates a GetServiceClassByNameBadRequest with default headers values
func NewGetServiceClassByNameBadRequest() *GetServiceClassByNameBadRequest {
	return &GetServiceClassByNameBadRequest{}
}

/*GetServiceClassByNameBadRequest handles this case with default header values.

Invalid name supplied
*/
type GetServiceClassByNameBadRequest struct {
	Payload *models.Error
}

func (o *GetServiceClassByNameBadRequest) Error() string {
	return fmt.Sprintf("[GET /serviceclass/{serviceClassName}][%d] getServiceClassByNameBadRequest  %+v", 400, o.Payload)
}

func (o *GetServiceClassByNameBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetServiceClassByNameNotFound creates a GetServiceClassByNameNotFound with default headers values
func NewGetServiceClassByNameNotFound() *GetServiceClassByNameNotFound {
	return &GetServiceClassByNameNotFound{}
}

/*GetServiceClassByNameNotFound handles this case with default header values.

Service class not found
*/
type GetServiceClassByNameNotFound struct {
	Payload *models.Error
}

func (o *GetServiceClassByNameNotFound) Error() string {
	return fmt.Sprintf("[GET /serviceclass/{serviceClassName}][%d] getServiceClassByNameNotFound  %+v", 404, o.Payload)
}

func (o *GetServiceClassByNameNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetServiceClassByNameDefault creates a GetServiceClassByNameDefault with default headers values
func NewGetServiceClassByNameDefault(code int) *GetServiceClassByNameDefault {
	return &GetServiceClassByNameDefault{
		_statusCode: code,
	}
}

/*GetServiceClassByNameDefault handles this case with default header values.

Generic error response
*/
type GetServiceClassByNameDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the get service class by name default response
func (o *GetServiceClassByNameDefault) Code() int {
	return o._statusCode
}

func (o *GetServiceClassByNameDefault) Error() string {
	return fmt.Sprintf("[GET /serviceclass/{serviceClassName}][%d] getServiceClassByName default  %+v", o._statusCode, o.Payload)
}

func (o *GetServiceClassByNameDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
