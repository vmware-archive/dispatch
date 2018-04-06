///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package serviceaccount

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware/dispatch/pkg/identity-manager/gen/models"
)

// AddServiceAccountReader is a Reader for the AddServiceAccount structure.
type AddServiceAccountReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *AddServiceAccountReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 201:
		result := NewAddServiceAccountCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewAddServiceAccountBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 409:
		result := NewAddServiceAccountConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 500:
		result := NewAddServiceAccountInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewAddServiceAccountDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewAddServiceAccountCreated creates a AddServiceAccountCreated with default headers values
func NewAddServiceAccountCreated() *AddServiceAccountCreated {
	return &AddServiceAccountCreated{}
}

/*AddServiceAccountCreated handles this case with default header values.

created
*/
type AddServiceAccountCreated struct {
	Payload *models.ServiceAccount
}

func (o *AddServiceAccountCreated) Error() string {
	return fmt.Sprintf("[POST /v1/iam/serviceaccount][%d] addServiceAccountCreated  %+v", 201, o.Payload)
}

func (o *AddServiceAccountCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ServiceAccount)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddServiceAccountBadRequest creates a AddServiceAccountBadRequest with default headers values
func NewAddServiceAccountBadRequest() *AddServiceAccountBadRequest {
	return &AddServiceAccountBadRequest{}
}

/*AddServiceAccountBadRequest handles this case with default header values.

Invalid input
*/
type AddServiceAccountBadRequest struct {
	Payload *models.Error
}

func (o *AddServiceAccountBadRequest) Error() string {
	return fmt.Sprintf("[POST /v1/iam/serviceaccount][%d] addServiceAccountBadRequest  %+v", 400, o.Payload)
}

func (o *AddServiceAccountBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddServiceAccountConflict creates a AddServiceAccountConflict with default headers values
func NewAddServiceAccountConflict() *AddServiceAccountConflict {
	return &AddServiceAccountConflict{}
}

/*AddServiceAccountConflict handles this case with default header values.

Already Exists
*/
type AddServiceAccountConflict struct {
	Payload *models.Error
}

func (o *AddServiceAccountConflict) Error() string {
	return fmt.Sprintf("[POST /v1/iam/serviceaccount][%d] addServiceAccountConflict  %+v", 409, o.Payload)
}

func (o *AddServiceAccountConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddServiceAccountInternalServerError creates a AddServiceAccountInternalServerError with default headers values
func NewAddServiceAccountInternalServerError() *AddServiceAccountInternalServerError {
	return &AddServiceAccountInternalServerError{}
}

/*AddServiceAccountInternalServerError handles this case with default header values.

Internal Error
*/
type AddServiceAccountInternalServerError struct {
	Payload *models.Error
}

func (o *AddServiceAccountInternalServerError) Error() string {
	return fmt.Sprintf("[POST /v1/iam/serviceaccount][%d] addServiceAccountInternalServerError  %+v", 500, o.Payload)
}

func (o *AddServiceAccountInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewAddServiceAccountDefault creates a AddServiceAccountDefault with default headers values
func NewAddServiceAccountDefault(code int) *AddServiceAccountDefault {
	return &AddServiceAccountDefault{
		_statusCode: code,
	}
}

/*AddServiceAccountDefault handles this case with default header values.

Generic error response
*/
type AddServiceAccountDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the add service account default response
func (o *AddServiceAccountDefault) Code() int {
	return o._statusCode
}

func (o *AddServiceAccountDefault) Error() string {
	return fmt.Sprintf("[POST /v1/iam/serviceaccount][%d] addServiceAccount default  %+v", o._statusCode, o.Payload)
}

func (o *AddServiceAccountDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
