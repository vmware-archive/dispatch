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

	case 401:
		result := NewHomeUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 403:
		result := NewHomeForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

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
	Payload *v1.Message
}

func (o *HomeOK) Error() string {
	return fmt.Sprintf("[GET /home][%d] homeOK  %+v", 200, o.Payload)
}

func (o *HomeOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Message)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewHomeUnauthorized creates a HomeUnauthorized with default headers values
func NewHomeUnauthorized() *HomeUnauthorized {
	return &HomeUnauthorized{}
}

/*HomeUnauthorized handles this case with default header values.

Unauthorized Request
*/
type HomeUnauthorized struct {
	Payload *v1.Error
}

func (o *HomeUnauthorized) Error() string {
	return fmt.Sprintf("[GET /home][%d] homeUnauthorized  %+v", 401, o.Payload)
}

func (o *HomeUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewHomeForbidden creates a HomeForbidden with default headers values
func NewHomeForbidden() *HomeForbidden {
	return &HomeForbidden{}
}

/*HomeForbidden handles this case with default header values.

access to this resource is forbidden
*/
type HomeForbidden struct {
	Payload *v1.Error
}

func (o *HomeForbidden) Error() string {
	return fmt.Sprintf("[GET /home][%d] homeForbidden  %+v", 403, o.Payload)
}

func (o *HomeForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

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

	Payload *v1.Error
}

// Code gets the status code for the home default response
func (o *HomeDefault) Code() int {
	return o._statusCode
}

func (o *HomeDefault) Error() string {
	return fmt.Sprintf("[GET /home][%d] home default  %+v", o._statusCode, o.Payload)
}

func (o *HomeDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
