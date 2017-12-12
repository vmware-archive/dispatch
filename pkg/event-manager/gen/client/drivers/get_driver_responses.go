///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package drivers

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

// GetDriverReader is a Reader for the GetDriver structure.
type GetDriverReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetDriverReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetDriverOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewGetDriverBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewGetDriverNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 500:
		result := NewGetDriverInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewGetDriverDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetDriverOK creates a GetDriverOK with default headers values
func NewGetDriverOK() *GetDriverOK {
	return &GetDriverOK{}
}

/*GetDriverOK handles this case with default header values.

Successful operation
*/
type GetDriverOK struct {
	Payload *models.Driver
}

func (o *GetDriverOK) Error() string {
	return fmt.Sprintf("[GET /drivers/{driverName}][%d] getDriverOK  %+v", 200, o.Payload)
}

func (o *GetDriverOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Driver)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetDriverBadRequest creates a GetDriverBadRequest with default headers values
func NewGetDriverBadRequest() *GetDriverBadRequest {
	return &GetDriverBadRequest{}
}

/*GetDriverBadRequest handles this case with default header values.

Invalid Name supplied
*/
type GetDriverBadRequest struct {
	Payload *models.Error
}

func (o *GetDriverBadRequest) Error() string {
	return fmt.Sprintf("[GET /drivers/{driverName}][%d] getDriverBadRequest  %+v", 400, o.Payload)
}

func (o *GetDriverBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetDriverNotFound creates a GetDriverNotFound with default headers values
func NewGetDriverNotFound() *GetDriverNotFound {
	return &GetDriverNotFound{}
}

/*GetDriverNotFound handles this case with default header values.

Driver not found
*/
type GetDriverNotFound struct {
	Payload *models.Error
}

func (o *GetDriverNotFound) Error() string {
	return fmt.Sprintf("[GET /drivers/{driverName}][%d] getDriverNotFound  %+v", 404, o.Payload)
}

func (o *GetDriverNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetDriverInternalServerError creates a GetDriverInternalServerError with default headers values
func NewGetDriverInternalServerError() *GetDriverInternalServerError {
	return &GetDriverInternalServerError{}
}

/*GetDriverInternalServerError handles this case with default header values.

Internal server error
*/
type GetDriverInternalServerError struct {
	Payload *models.Error
}

func (o *GetDriverInternalServerError) Error() string {
	return fmt.Sprintf("[GET /drivers/{driverName}][%d] getDriverInternalServerError  %+v", 500, o.Payload)
}

func (o *GetDriverInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetDriverDefault creates a GetDriverDefault with default headers values
func NewGetDriverDefault(code int) *GetDriverDefault {
	return &GetDriverDefault{
		_statusCode: code,
	}
}

/*GetDriverDefault handles this case with default header values.

Unknown error
*/
type GetDriverDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the get driver default response
func (o *GetDriverDefault) Code() int {
	return o._statusCode
}

func (o *GetDriverDefault) Error() string {
	return fmt.Sprintf("[GET /drivers/{driverName}][%d] getDriver default  %+v", o._statusCode, o.Payload)
}

func (o *GetDriverDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
