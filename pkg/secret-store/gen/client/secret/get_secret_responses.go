///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package secret

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// GetSecretReader is a Reader for the GetSecret structure.
type GetSecretReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetSecretReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetSecretOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewGetSecretBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewGetSecretNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewGetSecretDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetSecretOK creates a GetSecretOK with default headers values
func NewGetSecretOK() *GetSecretOK {
	return &GetSecretOK{}
}

/*GetSecretOK handles this case with default header values.

The secret identified by the secretName
*/
type GetSecretOK struct {
	Payload *v1.Secret
}

func (o *GetSecretOK) Error() string {
	return fmt.Sprintf("[GET /{secretName}][%d] getSecretOK  %+v", 200, o.Payload)
}

func (o *GetSecretOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Secret)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSecretBadRequest creates a GetSecretBadRequest with default headers values
func NewGetSecretBadRequest() *GetSecretBadRequest {
	return &GetSecretBadRequest{}
}

/*GetSecretBadRequest handles this case with default header values.

Bad Request
*/
type GetSecretBadRequest struct {
	Payload *v1.Error
}

func (o *GetSecretBadRequest) Error() string {
	return fmt.Sprintf("[GET /{secretName}][%d] getSecretBadRequest  %+v", 400, o.Payload)
}

func (o *GetSecretBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSecretNotFound creates a GetSecretNotFound with default headers values
func NewGetSecretNotFound() *GetSecretNotFound {
	return &GetSecretNotFound{}
}

/*GetSecretNotFound handles this case with default header values.

Resource Not Found if no secret exists with the given name
*/
type GetSecretNotFound struct {
	Payload *v1.Error
}

func (o *GetSecretNotFound) Error() string {
	return fmt.Sprintf("[GET /{secretName}][%d] getSecretNotFound  %+v", 404, o.Payload)
}

func (o *GetSecretNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSecretDefault creates a GetSecretDefault with default headers values
func NewGetSecretDefault(code int) *GetSecretDefault {
	return &GetSecretDefault{
		_statusCode: code,
	}
}

/*GetSecretDefault handles this case with default header values.

Standard error
*/
type GetSecretDefault struct {
	_statusCode int

	Payload *v1.Error
}

// Code gets the status code for the get secret default response
func (o *GetSecretDefault) Code() int {
	return o._statusCode
}

func (o *GetSecretDefault) Error() string {
	return fmt.Sprintf("[GET /{secretName}][%d] getSecret default  %+v", o._statusCode, o.Payload)
}

func (o *GetSecretDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
