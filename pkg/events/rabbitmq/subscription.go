///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package rabbitmq

// NO TESTS

// subscription is an implementation of Subscription interface, used to
// stop not needed subscriptions
type subscription struct {
	done         chan struct{}
	unsubscribed bool
}

func (sub *subscription) Unsubscribe() error {
	if !sub.unsubscribed {
		close(sub.done)
		sub.unsubscribed = true
	}
	return nil
}
