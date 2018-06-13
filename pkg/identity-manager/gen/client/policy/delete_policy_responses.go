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

	"github.com/vmware/dispatch/pkg/api/v1"
)

// DeletePolicyReader is a Reader for the DeletePolicy structure.
type DeletePolicyReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeletePolicyReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewDeletePolicyOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 400:
		result := NewDeletePolicyBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 401:
		result := NewDeletePolicyUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 403:
		result := NewDeletePolicyForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewDeletePolicyNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewDeletePolicyDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewDeletePolicyOK creates a DeletePolicyOK with default headers values
func NewDeletePolicyOK() *DeletePolicyOK {
	return &DeletePolicyOK{}
}

/*DeletePolicyOK handles this case with default header values.

Successful operation
*/
type DeletePolicyOK struct {
	Payload *v1.Policy
}

func (o *DeletePolicyOK) Error() string {
	return fmt.Sprintf("[DELETE /v1/iam/policy/{policyName}][%d] deletePolicyOK  %+v", 200, o.Payload)
}

func (o *DeletePolicyOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Policy)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeletePolicyBadRequest creates a DeletePolicyBadRequest with default headers values
func NewDeletePolicyBadRequest() *DeletePolicyBadRequest {
	return &DeletePolicyBadRequest{}
}

/*DeletePolicyBadRequest handles this case with default header values.

Invalid Name supplied
*/
type DeletePolicyBadRequest struct {
	Payload *v1.Error
}

func (o *DeletePolicyBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /v1/iam/policy/{policyName}][%d] deletePolicyBadRequest  %+v", 400, o.Payload)
}

func (o *DeletePolicyBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeletePolicyUnauthorized creates a DeletePolicyUnauthorized with default headers values
func NewDeletePolicyUnauthorized() *DeletePolicyUnauthorized {
	return &DeletePolicyUnauthorized{}
}

/*DeletePolicyUnauthorized handles this case with default header values.

Unauthorized Request
*/
type DeletePolicyUnauthorized struct {
	Payload *v1.Error
}

func (o *DeletePolicyUnauthorized) Error() string {
	return fmt.Sprintf("[DELETE /v1/iam/policy/{policyName}][%d] deletePolicyUnauthorized  %+v", 401, o.Payload)
}

func (o *DeletePolicyUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeletePolicyForbidden creates a DeletePolicyForbidden with default headers values
func NewDeletePolicyForbidden() *DeletePolicyForbidden {
	return &DeletePolicyForbidden{}
}

/*DeletePolicyForbidden handles this case with default header values.

access to this resource is forbidden
*/
type DeletePolicyForbidden struct {
	Payload *v1.Error
}

func (o *DeletePolicyForbidden) Error() string {
	return fmt.Sprintf("[DELETE /v1/iam/policy/{policyName}][%d] deletePolicyForbidden  %+v", 403, o.Payload)
}

func (o *DeletePolicyForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeletePolicyNotFound creates a DeletePolicyNotFound with default headers values
func NewDeletePolicyNotFound() *DeletePolicyNotFound {
	return &DeletePolicyNotFound{}
}

/*DeletePolicyNotFound handles this case with default header values.

Policy not found
*/
type DeletePolicyNotFound struct {
	Payload *v1.Error
}

func (o *DeletePolicyNotFound) Error() string {
	return fmt.Sprintf("[DELETE /v1/iam/policy/{policyName}][%d] deletePolicyNotFound  %+v", 404, o.Payload)
}

func (o *DeletePolicyNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeletePolicyDefault creates a DeletePolicyDefault with default headers values
func NewDeletePolicyDefault(code int) *DeletePolicyDefault {
	return &DeletePolicyDefault{
		_statusCode: code,
	}
}

/*DeletePolicyDefault handles this case with default header values.

Unknown error
*/
type DeletePolicyDefault struct {
	_statusCode int

	Payload *v1.Error
}

// Code gets the status code for the delete policy default response
func (o *DeletePolicyDefault) Code() int {
	return o._statusCode
}

func (o *DeletePolicyDefault) Error() string {
	return fmt.Sprintf("[DELETE /v1/iam/policy/{policyName}][%d] deletePolicy default  %+v", o._statusCode, o.Payload)
}

func (o *DeletePolicyDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(v1.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
