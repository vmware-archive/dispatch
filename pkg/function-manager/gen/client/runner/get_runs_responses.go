///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package runner

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/vmware/dispatch/pkg/function-manager/gen/models"
)

// GetRunsReader is a Reader for the GetRuns structure.
type GetRunsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetRunsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetRunsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 404:
		result := NewGetRunsNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 500:
		result := NewGetRunsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetRunsOK creates a GetRunsOK with default headers values
func NewGetRunsOK() *GetRunsOK {
	return &GetRunsOK{}
}

/*GetRunsOK handles this case with default header values.

List of function runs
*/
type GetRunsOK struct {
	Payload models.GetRunsOKBody
}

func (o *GetRunsOK) Error() string {
	return fmt.Sprintf("[GET /runs][%d] getRunsOK  %+v", 200, o.Payload)
}

func (o *GetRunsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetRunsNotFound creates a GetRunsNotFound with default headers values
func NewGetRunsNotFound() *GetRunsNotFound {
	return &GetRunsNotFound{}
}

/*GetRunsNotFound handles this case with default header values.

Function not found
*/
type GetRunsNotFound struct {
	Payload *models.Error
}

func (o *GetRunsNotFound) Error() string {
	return fmt.Sprintf("[GET /runs][%d] getRunsNotFound  %+v", 404, o.Payload)
}

func (o *GetRunsNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetRunsInternalServerError creates a GetRunsInternalServerError with default headers values
func NewGetRunsInternalServerError() *GetRunsInternalServerError {
	return &GetRunsInternalServerError{}
}

/*GetRunsInternalServerError handles this case with default header values.

Internal error
*/
type GetRunsInternalServerError struct {
	Payload *models.Error
}

func (o *GetRunsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /runs][%d] getRunsInternalServerError  %+v", 500, o.Payload)
}

func (o *GetRunsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
