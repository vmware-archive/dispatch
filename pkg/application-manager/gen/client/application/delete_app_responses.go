///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package application

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// DeleteAppReader is a Reader for the DeleteApp structure.
type DeleteAppReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteAppReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewDeleteAppOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewDeleteAppBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewDeleteAppNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 500:
		result := NewDeleteAppInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewDeleteAppOK creates a DeleteAppOK with default headers values
func NewDeleteAppOK() *DeleteAppOK {
	return &DeleteAppOK{}
}

/*DeleteAppOK handles this case with default header values.

Successful operation
*/
type DeleteAppOK struct {
	Payload *v1.Application
}

func (o *DeleteAppOK) Error() string {
	return fmt.Sprintf("[DELETE /{application}][%d] deleteAppOK  %+v", 200, o.Payload)
}

func (o *DeleteAppOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Application)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteAppBadRequest creates a DeleteAppBadRequest with default headers values
func NewDeleteAppBadRequest() *DeleteAppBadRequest {
	return &DeleteAppBadRequest{}
}

/*DeleteAppBadRequest handles this case with default header values.

Invalid Name supplied
*/
type DeleteAppBadRequest struct {
	Payload *v1.Error
}

func (o *DeleteAppBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /{application}][%d] deleteAppBadRequest  %+v", 400, o.Payload)
}

func (o *DeleteAppBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteAppNotFound creates a DeleteAppNotFound with default headers values
func NewDeleteAppNotFound() *DeleteAppNotFound {
	return &DeleteAppNotFound{}
}

/*DeleteAppNotFound handles this case with default header values.

Application not found
*/
type DeleteAppNotFound struct {
	Payload *v1.Error
}

func (o *DeleteAppNotFound) Error() string {
	return fmt.Sprintf("[DELETE /{application}][%d] deleteAppNotFound  %+v", 404, o.Payload)
}

func (o *DeleteAppNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteAppInternalServerError creates a DeleteAppInternalServerError with default headers values
func NewDeleteAppInternalServerError() *DeleteAppInternalServerError {
	return &DeleteAppInternalServerError{}
}

/*DeleteAppInternalServerError handles this case with default header values.

Internal error
*/
type DeleteAppInternalServerError struct {
	Payload *v1.Error
}

func (o *DeleteAppInternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /{application}][%d] deleteAppInternalServerError  %+v", 500, o.Payload)
}

func (o *DeleteAppInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
