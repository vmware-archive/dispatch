///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

const (
	errReasonDeploymentNotFound      = "DeploymentNotFound"
	errReasonDeploymentAlreadyExists = "DeploymentAlreadyExists"
	errReasonDeploymentNotAvaialble  = "DeploymentNotAvailable"
	errReasonUnknown                 = "Unknown"
)

// Causer defines interface for eventdriver error
type Causer interface {
	Cause() error
}

// EventdriverErrorDeploymentNotFound defines deployment not found error for event driver
type EventdriverErrorDeploymentNotFound struct {
	Err error `json:"err"`
}

func (err *EventdriverErrorDeploymentNotFound) Error() string {
	return errReasonDeploymentNotFound
}

// Cause returns the underlying cause error
func (err *EventdriverErrorDeploymentNotFound) Cause() error {
	return err.Err
}

// EventdriverErrorDeploymentNotAvaialble defines deployment not available error for event driver
type EventdriverErrorDeploymentNotAvaialble struct {
	Err error `json:"err"`
}

func (err *EventdriverErrorDeploymentNotAvaialble) Error() string {
	return errReasonDeploymentNotAvaialble
}

// Cause returns the underlying cause error
func (err *EventdriverErrorDeploymentNotAvaialble) Cause() error {
	return err.Err
}

// EventdriverErrorDeploymentAlreadyExists defines deployment already exists error for event driver
type EventdriverErrorDeploymentAlreadyExists struct {
	Err error `json:"err"`
}

func (err *EventdriverErrorDeploymentAlreadyExists) Error() string {
	return errReasonDeploymentAlreadyExists
}

// Cause returns the underlying cause error
func (err *EventdriverErrorDeploymentAlreadyExists) Cause() error {
	return err.Err
}

// EventdriverErrorUnknown defines unknonwn error for event driver
type EventdriverErrorUnknown struct {
	Err error `json:"err"`
}

func (err *EventdriverErrorUnknown) Error() string {
	return errReasonUnknown
}

// Cause returns the underlying cause error
func (err *EventdriverErrorUnknown) Cause() error {
	return err.Err
}
