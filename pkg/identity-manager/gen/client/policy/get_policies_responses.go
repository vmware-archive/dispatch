///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package policy

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware/dispatch/pkg/identity-manager/gen/models"
)

// GetPoliciesReader is a Reader for the GetPolicies structure.
type GetPoliciesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetPoliciesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetPoliciesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 500:
		result := NewGetPoliciesInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewGetPoliciesDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetPoliciesOK creates a GetPoliciesOK with default headers values
func NewGetPoliciesOK() *GetPoliciesOK {
	return &GetPoliciesOK{}
}

/*GetPoliciesOK handles this case with default header values.

Successful operation
*/
type GetPoliciesOK struct {
	Payload []*models.Policy
}

func (o *GetPoliciesOK) Error() string {
	return fmt.Sprintf("[GET /v1/iam/policy][%d] getPoliciesOK  %+v", 200, o.Payload)
}

func (o *GetPoliciesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPoliciesInternalServerError creates a GetPoliciesInternalServerError with default headers values
func NewGetPoliciesInternalServerError() *GetPoliciesInternalServerError {
	return &GetPoliciesInternalServerError{}
}

/*GetPoliciesInternalServerError handles this case with default header values.

Internal Error
*/
type GetPoliciesInternalServerError struct {
	Payload *models.Error
}

func (o *GetPoliciesInternalServerError) Error() string {
	return fmt.Sprintf("[GET /v1/iam/policy][%d] getPoliciesInternalServerError  %+v", 500, o.Payload)
}

func (o *GetPoliciesInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPoliciesDefault creates a GetPoliciesDefault with default headers values
func NewGetPoliciesDefault(code int) *GetPoliciesDefault {
	return &GetPoliciesDefault{
		_statusCode: code,
	}
}

/*GetPoliciesDefault handles this case with default header values.

Unexpected Error
*/
type GetPoliciesDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the get policies default response
func (o *GetPoliciesDefault) Code() int {
	return o._statusCode
}

func (o *GetPoliciesDefault) Error() string {
	return fmt.Sprintf("[GET /v1/iam/policy][%d] getPolicies default  %+v", o._statusCode, o.Payload)
}

func (o *GetPoliciesDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
