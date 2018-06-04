// Code generated by go-swagger; DO NOT EDIT.

package organization

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/vmware/dispatch/pkg/api/v1"
)

// NewUpdateOrganizationParams creates a new UpdateOrganizationParams object
// no default values defined in spec.
func NewUpdateOrganizationParams() UpdateOrganizationParams {

	return UpdateOrganizationParams{}
}

// UpdateOrganizationParams contains all the bound params for the update organization operation
// typically these are obtained from a http.Request
//
// swagger:parameters updateOrganization
type UpdateOrganizationParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Organization object
	  Required: true
	  In: body
	*/
	Body *v1.Organization
	/*Name of Organization to work on
	  Required: true
	  Pattern: ^[\w\d\-]+$
	  In: path
	*/
	OrganizationName string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewUpdateOrganizationParams() beforehand.
func (o *UpdateOrganizationParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body v1.Organization
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("body", "body"))
			} else {
				res = append(res, errors.NewParseError("body", "body", "", err))
			}
		} else {

			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Body = &body
			}
		}
	} else {
		res = append(res, errors.Required("body", "body"))
	}
	rOrganizationName, rhkOrganizationName, _ := route.Params.GetOK("organizationName")
	if err := o.bindOrganizationName(rOrganizationName, rhkOrganizationName, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *UpdateOrganizationParams) bindOrganizationName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.OrganizationName = raw

	if err := o.validateOrganizationName(formats); err != nil {
		return err
	}

	return nil
}

func (o *UpdateOrganizationParams) validateOrganizationName(formats strfmt.Registry) error {

	if err := validate.Pattern("organizationName", "path", o.OrganizationName, `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}
