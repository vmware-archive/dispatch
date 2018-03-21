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

// RedirectReader is a Reader for the Redirect structure.
type RedirectReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *RedirectReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 302:
		result := NewRedirectFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewRedirectDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewRedirectFound creates a RedirectFound with default headers values
func NewRedirectFound() *RedirectFound {
	return &RedirectFound{}
}

/*RedirectFound handles this case with default header values.

redirect
*/
type RedirectFound struct {
	/*redirect location
	 */
	Location string
}

func (o *RedirectFound) Error() string {
	return fmt.Sprintf("[GET /v1/iam/redirect][%d] redirectFound ", 302)
}

func (o *RedirectFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response header Location
	o.Location = response.GetHeader("Location")

	return nil
}

// NewRedirectDefault creates a RedirectDefault with default headers values
func NewRedirectDefault(code int) *RedirectDefault {
	return &RedirectDefault{
		_statusCode: code,
	}
}

/*RedirectDefault handles this case with default header values.

error
*/
type RedirectDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the redirect default response
func (o *RedirectDefault) Code() int {
	return o._statusCode
}

func (o *RedirectDefault) Error() string {
	return fmt.Sprintf("[GET /v1/iam/redirect][%d] redirect default  %+v", o._statusCode, o.Payload)
}

func (o *RedirectDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
