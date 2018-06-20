///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package transport

// NO TESTS

// subscription is an implementation of Subscription interface, used to
// stop not needed subscriptions
type subscription struct {
	done         chan struct{}
	unsubscribed bool
	topic        string
	organization string
}

func (sub *subscription) Unsubscribe() error {
	if !sub.unsubscribed {
		close(sub.done)
		sub.unsubscribed = true
	}
	return nil
}

func (sub *subscription) GetTopic() string {
	return sub.topic
}

func (sub *subscription) GetOrganization() string {
	return sub.organization
}
