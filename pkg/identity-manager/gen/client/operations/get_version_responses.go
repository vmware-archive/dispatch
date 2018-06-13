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

	"github.com/vmware/dispatch/pkg/api/v1"
)

// GetVersionReader is a Reader for the GetVersion structure.
type GetVersionReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetVersionReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetVersionOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 401:
		result := NewGetVersionUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 403:
		result := NewGetVersionForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewGetVersionDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetVersionOK creates a GetVersionOK with default headers values
func NewGetVersionOK() *GetVersionOK {
	return &GetVersionOK{}
}

/*GetVersionOK handles this case with default header values.

version info
*/
type GetVersionOK struct {
	Payload *v1.Version
}

func (o *GetVersionOK) Error() string {
	return fmt.Sprintf("[GET /v1/version][%d] getVersionOK  %+v", 200, o.Payload)
}

func (o *GetVersionOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Version)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetVersionUnauthorized creates a GetVersionUnauthorized with default headers values
func NewGetVersionUnauthorized() *GetVersionUnauthorized {
	return &GetVersionUnauthorized{}
}

/*GetVersionUnauthorized handles this case with default header values.

Unauthorized Request
*/
type GetVersionUnauthorized struct {
	Payload *v1.Error
}

func (o *GetVersionUnauthorized) Error() string {
	return fmt.Sprintf("[GET /v1/version][%d] getVersionUnauthorized  %+v", 401, o.Payload)
}

func (o *GetVersionUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetVersionForbidden creates a GetVersionForbidden with default headers values
func NewGetVersionForbidden() *GetVersionForbidden {
	return &GetVersionForbidden{}
}

/*GetVersionForbidden handles this case with default header values.

access to this resource is forbidden
*/
type GetVersionForbidden struct {
	Payload *v1.Error
}

func (o *GetVersionForbidden) Error() string {
	return fmt.Sprintf("[GET /v1/version][%d] getVersionForbidden  %+v", 403, o.Payload)
}

func (o *GetVersionForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetVersionDefault creates a GetVersionDefault with default headers values
func NewGetVersionDefault(code int) *GetVersionDefault {
	return &GetVersionDefault{
		_statusCode: code,
	}
}

/*GetVersionDefault handles this case with default header values.

error
*/
type GetVersionDefault struct {
	_statusCode int

	Payload *v1.Error
}

// Code gets the status code for the get version default response
func (o *GetVersionDefault) Code() int {
	return o._statusCode
}

func (o *GetVersionDefault) Error() string {
	return fmt.Sprintf("[GET /v1/version][%d] getVersion default  %+v", o._statusCode, o.Payload)
}

func (o *GetVersionDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
