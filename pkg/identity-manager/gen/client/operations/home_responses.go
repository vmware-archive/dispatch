///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware/dispatch/pkg/identity-manager/gen/models"
)

// HomeReader is a Reader for the Home structure.
type HomeReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *HomeReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewHomeOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewHomeDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewHomeOK creates a HomeOK with default headers values
func NewHomeOK() *HomeOK {
	return &HomeOK{}
}

/*HomeOK handles this case with default header values.

home page
*/
type HomeOK struct {
	Payload *models.Message
}

func (o *HomeOK) Error() string {
	return fmt.Sprintf("[GET /v1/iam/home][%d] homeOK  %+v", 200, o.Payload)
}

func (o *HomeOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Message)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewHomeDefault creates a HomeDefault with default headers values
func NewHomeDefault(code int) *HomeDefault {
	return &HomeDefault{
		_statusCode: code,
	}
}

/*HomeDefault handles this case with default header values.

error
*/
type HomeDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the home default response
func (o *HomeDefault) Code() int {
	return o._statusCode
}

func (o *HomeDefault) Error() string {
	return fmt.Sprintf("[GET /v1/iam/home][%d] home default  %+v", o._statusCode, o.Payload)
}

func (o *HomeDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
