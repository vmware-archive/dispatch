///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package validator

import (
	"github.com/vmware/dispatch/pkg/events"
	govalidator "gopkg.in/go-playground/validator.v9"
)

var extraValidators = map[string]govalidator.Func{
	"eventtype": eventType,
}

func eventType(fl govalidator.FieldLevel) bool {
	return eventTypeRegex.MatchString(fl.Field().String())
}

var validator = NewDefaultValidator()

// Validate validates cloud event using default validator
func Validate(event *events.CloudEvent) error {
	return validator.Validate(event)
}

// defaultValidator implements default validator based on go-playground/validator
type defaultValidator struct {
	instance *govalidator.Validate
}

// NewDefaultValidator returns instance of events.Validator based on go-playground/validator
func NewDefaultValidator() events.Validator {
	instance := govalidator.New()
	for tag, f := range extraValidators {
		instance.RegisterValidation(tag, f)
	}

	return &defaultValidator{
		instance: instance,
	}
}

// Validate validates the event
func (v *defaultValidator) Validate(event *events.CloudEvent) error {
	return v.instance.Struct(event)
}
