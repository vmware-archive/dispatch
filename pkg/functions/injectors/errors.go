///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package injectors

type injectorError struct {
	Err error `json:"err"`
}

func (err *injectorError) Error() string {
	return err.Err.Error()
}

func (err *injectorError) AsUserErrorObject() interface{} {
	return err
}
