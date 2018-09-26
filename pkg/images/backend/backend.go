///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package backend

import (
	"context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/errors"
)

// Backend is the interface for image manager backend, like Knative
type Backend interface {
	AddImage(ctx context.Context, image *v1.Image) (*v1.Image, *errors.DispatchError)
	GetImage(ctx context.Context, meta *v1.Meta) (*v1.Image, *errors.DispatchError)
	DeleteImage(ctx context.Context, meta *v1.Meta) *errors.DispatchError
	ListImage(ctx context.Context, meta *v1.Meta) ([]*v1.Image, *errors.DispatchError)
	UpdateImage(ctx context.Context, image *v1.Image) (*v1.Image, *errors.DispatchError)
}
