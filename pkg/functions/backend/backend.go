///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"

	"github.com/vmware/dispatch/pkg/api/v1"
)

//Backend is the interface for a function-manager backend, such as Knative
type Backend interface {
	Add(ctx context.Context, function *v1.Function) (*v1.Function, error)
	Get(ctx context.Context, meta *v1.Meta) (*v1.Function, error)
	Delete(ctx context.Context, meta *v1.Meta) error
	List(ctx context.Context, meta *v1.Meta) ([]*v1.Function, error)
	Update(ctx context.Context, function *v1.Function) (*v1.Function, error)

	RunEndpoint(ctx context.Context, meta *v1.Meta) (string, string, error)
}

//NotFound is a typed error meaning that the requested entity was not found
type NotFound struct {
	error
}

//Cause returns the parent error
func (nf NotFound) Cause() error {
	return nf.error
}

//AlreadyExists is a typed error meaning that the entity being persisted already exists
type AlreadyExists struct {
	error
}

//Cause returns the parent error
func (ae AlreadyExists) Cause() error {
	return ae.error
}

//ValidationError is a typed error meaning that the entity being persisted is invalid
type ValidationError struct {
	error
}

//Cause returns the parent error
func (ve ValidationError) Cause() error {
	return ve.error
}
