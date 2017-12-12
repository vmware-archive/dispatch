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

// DeleteDriverReader is a Reader for the DeleteDriver structure.
type DeleteDriverReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteDriverReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewDeleteDriverOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewDeleteDriverBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewDeleteDriverNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 500:
		result := NewDeleteDriverInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewDeleteDriverDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewDeleteDriverOK creates a DeleteDriverOK with default headers values
func NewDeleteDriverOK() *DeleteDriverOK {
	return &DeleteDriverOK{}
}

/*DeleteDriverOK handles this case with default header values.

successful operation
*/
type DeleteDriverOK struct {
	Payload *models.Driver
}

func (o *DeleteDriverOK) Error() string {
	return fmt.Sprintf("[DELETE /drivers/{driverName}][%d] deleteDriverOK  %+v", 200, o.Payload)
}

func (o *DeleteDriverOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Driver)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteDriverBadRequest creates a DeleteDriverBadRequest with default headers values
func NewDeleteDriverBadRequest() *DeleteDriverBadRequest {
	return &DeleteDriverBadRequest{}
}

/*DeleteDriverBadRequest handles this case with default header values.

Invalid ID supplied
*/
type DeleteDriverBadRequest struct {
	Payload *models.Error
}

func (o *DeleteDriverBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /drivers/{driverName}][%d] deleteDriverBadRequest  %+v", 400, o.Payload)
}

func (o *DeleteDriverBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteDriverNotFound creates a DeleteDriverNotFound with default headers values
func NewDeleteDriverNotFound() *DeleteDriverNotFound {
	return &DeleteDriverNotFound{}
}

/*DeleteDriverNotFound handles this case with default header values.

Driver not found
*/
type DeleteDriverNotFound struct {
	Payload *models.Error
}

func (o *DeleteDriverNotFound) Error() string {
	return fmt.Sprintf("[DELETE /drivers/{driverName}][%d] deleteDriverNotFound  %+v", 404, o.Payload)
}

func (o *DeleteDriverNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteDriverInternalServerError creates a DeleteDriverInternalServerError with default headers values
func NewDeleteDriverInternalServerError() *DeleteDriverInternalServerError {
	return &DeleteDriverInternalServerError{}
}

/*DeleteDriverInternalServerError handles this case with default header values.

Internal server error
*/
type DeleteDriverInternalServerError struct {
	Payload *models.Error
}

func (o *DeleteDriverInternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /drivers/{driverName}][%d] deleteDriverInternalServerError  %+v", 500, o.Payload)
}

func (o *DeleteDriverInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteDriverDefault creates a DeleteDriverDefault with default headers values
func NewDeleteDriverDefault(code int) *DeleteDriverDefault {
	return &DeleteDriverDefault{
		_statusCode: code,
	}
}

/*DeleteDriverDefault handles this case with default header values.

Generic error response
*/
type DeleteDriverDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the delete driver default response
func (o *DeleteDriverDefault) Code() int {
	return o._statusCode
}

func (o *DeleteDriverDefault) Error() string {
	return fmt.Sprintf("[DELETE /drivers/{driverName}][%d] deleteDriver default  %+v", o._statusCode, o.Payload)
}

func (o *DeleteDriverDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
