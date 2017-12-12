///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////// Code generated by go-swagger; DO NOT EDIT.

package base_image

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"errors"
	"net/url"
	golangswaggerpaths "path"
	"strings"
)

// DeleteBaseImageByNameURL generates an URL for the delete base image by name operation
type DeleteBaseImageByNameURL struct {
	BaseImageName string

	_basePath string
	// avoid unkeyed usage
	_ struct{}
}

// WithBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *DeleteBaseImageByNameURL) WithBasePath(bp string) *DeleteBaseImageByNameURL {
	o.SetBasePath(bp)
	return o
}

// SetBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *DeleteBaseImageByNameURL) SetBasePath(bp string) {
	o._basePath = bp
}

// Build a url path and query string
func (o *DeleteBaseImageByNameURL) Build() (*url.URL, error) {
	var result url.URL

	var _path = "/base/{baseImageName}"

	baseImageName := o.BaseImageName
	if baseImageName != "" {
		_path = strings.Replace(_path, "{baseImageName}", baseImageName, -1)
	} else {
		return nil, errors.New("BaseImageName is required on DeleteBaseImageByNameURL")
	}
	_basePath := o._basePath
	if _basePath == "" {
		_basePath = "/v1/image"
	}
	result.Path = golangswaggerpaths.Join(_basePath, _path)

	return &result, nil
}

// Must is a helper function to panic when the url builder returns an error
func (o *DeleteBaseImageByNameURL) Must(u *url.URL, err error) *url.URL {
	if err != nil {
		panic(err)
	}
	if u == nil {
		panic("url can't be nil")
	}
	return u
}

// String returns the string representation of the path with query string
func (o *DeleteBaseImageByNameURL) String() string {
	return o.Must(o.Build()).String()
}

// BuildFull builds a full url with scheme, host, path and query string
func (o *DeleteBaseImageByNameURL) BuildFull(scheme, host string) (*url.URL, error) {
	if scheme == "" {
		return nil, errors.New("scheme is required for a full url on DeleteBaseImageByNameURL")
	}
	if host == "" {
		return nil, errors.New("host is required for a full url on DeleteBaseImageByNameURL")
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
func (o *DeleteBaseImageByNameURL) StringFull(scheme, host string) string {
	return o.Must(o.BuildFull(scheme, host)).String()
}
