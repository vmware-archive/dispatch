///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

// Code generated by go-swagger; DO NOT EDIT.

package secret

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"errors"
	"net/url"
	golangswaggerpaths "path"
	"strings"

	"github.com/go-openapi/swag"
)

// DeleteSecretURL generates an URL for the delete secret operation
type DeleteSecretURL struct {
	SecretName string

	Tags []string

	_basePath string
	// avoid unkeyed usage
	_ struct{}
}

// WithBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *DeleteSecretURL) WithBasePath(bp string) *DeleteSecretURL {
	o.SetBasePath(bp)
	return o
}

// SetBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *DeleteSecretURL) SetBasePath(bp string) {
	o._basePath = bp
}

// Build a url path and query string
func (o *DeleteSecretURL) Build() (*url.URL, error) {
	var result url.URL

	var _path = "/{secretName}"

	secretName := o.SecretName
	if secretName != "" {
		_path = strings.Replace(_path, "{secretName}", secretName, -1)
	} else {
		return nil, errors.New("SecretName is required on DeleteSecretURL")
	}

	_basePath := o._basePath
	if _basePath == "" {
		_basePath = "/v1/secret"
	}
	result.Path = golangswaggerpaths.Join(_basePath, _path)

	qs := make(url.Values)

	var tagsIR []string
	for _, tagsI := range o.Tags {
		tagsIS := tagsI
		if tagsIS != "" {
			tagsIR = append(tagsIR, tagsIS)
		}
	}

	tags := swag.JoinByFormat(tagsIR, "multi")

	for _, qsv := range tags {
		qs.Add("tags", qsv)
	}

	result.RawQuery = qs.Encode()

	return &result, nil
}

// Must is a helper function to panic when the url builder returns an error
func (o *DeleteSecretURL) Must(u *url.URL, err error) *url.URL {
	if err != nil {
		panic(err)
	}
	if u == nil {
		panic("url can't be nil")
	}
	return u
}

// String returns the string representation of the path with query string
func (o *DeleteSecretURL) String() string {
	return o.Must(o.Build()).String()
}

// BuildFull builds a full url with scheme, host, path and query string
func (o *DeleteSecretURL) BuildFull(scheme, host string) (*url.URL, error) {
	if scheme == "" {
		return nil, errors.New("scheme is required for a full url on DeleteSecretURL")
	}
	if host == "" {
		return nil, errors.New("host is required for a full url on DeleteSecretURL")
	}

	base, err := o.Build()
	if err != nil {
		return nil, err
	}

	base.Scheme = scheme
	base.Host = host
	return base, nil
}

// StringFull returns the string representation of a complete url
func (o *DeleteSecretURL) StringFull(scheme, host string) string {
	return o.Must(o.BuildFull(scheme, host)).String()
}
