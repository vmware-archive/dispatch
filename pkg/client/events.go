///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client

import (
	"context"
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/vmware/dispatch/pkg/api/v1"
	swaggerclient "github.com/vmware/dispatch/pkg/event-manager/gen/client"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/events"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/subscriptions"
)

// EventsClient defines the event client interface
type EventsClient interface {
	// Emit an event
	EmitEvent(ctx context.Context, organizationID string, emission *v1.Emission) (*v1.Emission, error)

	// Subscriptions
	CreateSubscription(ctx context.Context, organizationID string, subscription *v1.Subscription) (*v1.Subscription, error)
	DeleteSubscription(ctx context.Context, organizationID string, subscriptionName string) (*v1.Subscription, error)
	GetSubscription(ctx context.Context, organizationID string, subscriptionName string) (*v1.Subscription, error)
	ListSubscriptions(ctx context.Context, organizationID string) ([]v1.Subscription, error)
	UpdateSubscription(ctx context.Context, organizationID string, subscription *v1.Subscription) (*v1.Subscription, error)

	// Event Drivers
	CreateEventDriver(ctx context.Context, organizationID string, eventDriver *v1.EventDriver) (*v1.EventDriver, error)
	DeleteEventDriver(ctx context.Context, organizationID string, eventDriverName string) (*v1.EventDriver, error)
	GetEventDriver(ctx context.Context, organizationID string, eventDriverName string) (*v1.EventDriver, error)
	ListEventDrivers(ctx context.Context, organizationID string) ([]v1.EventDriver, error)
	UpdateEventDriver(ctx context.Context, organizationID string, eventDriver *v1.EventDriver) (*v1.EventDriver, error)

	// Event Driver Types
	CreateEventDriverType(ctx context.Context, organizationID string, eventDriverType *v1.EventDriverType) (*v1.EventDriverType, error)
	DeleteEventDriverType(ctx context.Context, organizationID string, eventDriverTypeName string) (*v1.EventDriverType, error)
	GetEventDriverType(ctx context.Context, organizationID string, eventDriverTypeName string) (*v1.EventDriverType, error)
	ListEventDriverTypes(ctx context.Context, organizationID string) ([]v1.EventDriverType, error)
	UpdateEventDriverType(ctx context.Context, organizationID string, eventDriverType *v1.EventDriverType) (*v1.EventDriverType, error)
}

// DefaultEventsClient defines the default client for events API
type DefaultEventsClient struct {
	baseClient

	client *swaggerclient.EventManager
	auth   runtime.ClientAuthInfoWriter
}

// NewEventsClient is used to create a new subscriptions client
func NewEventsClient(host string, auth runtime.ClientAuthInfoWriter, organizationID string) *DefaultEventsClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultEventsClient{
		baseClient: baseClient{
			organizationID: organizationID,
		},
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// EmitEvent emits an event
func (c *DefaultEventsClient) EmitEvent(ctx context.Context, organizationID string, emission *v1.Emission) (*v1.Emission, error) {
	params := events.EmitEventParams{
		Body:         emission,
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Events.EmitEvent(&params, c.auth)
	if err != nil {
		return nil, emitEventSwaggerError(err)
	}
	return response.Payload, nil
}

func emitEventSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *events.EmitEventBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *events.EmitEventUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *events.EmitEventForbidden:
		return NewErrorForbidden(v.Payload)
	case *events.EmitEventDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// CreateSubscription creates and adds a new subscription
func (c *DefaultEventsClient) CreateSubscription(ctx context.Context, organizationID string, subscription *v1.Subscription) (*v1.Subscription, error) {
	params := subscriptions.AddSubscriptionParams{
		Context:      ctx,
		Body:         subscription,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Subscriptions.AddSubscription(&params, c.auth)
	if err != nil {
		return nil, createSubscriptionSwaggerError(err)
	}
	return response.Payload, nil
}

func createSubscriptionSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *subscriptions.AddSubscriptionBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *subscriptions.AddSubscriptionUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *subscriptions.AddSubscriptionForbidden:
		return NewErrorForbidden(v.Payload)
	case *subscriptions.AddSubscriptionConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *subscriptions.AddSubscriptionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteSubscription deletes a subscription
func (c *DefaultEventsClient) DeleteSubscription(ctx context.Context, organizationID string, subscriptionName string) (*v1.Subscription, error) {
	params := subscriptions.DeleteSubscriptionParams{
		Context:          ctx,
		SubscriptionName: subscriptionName,
		XDispatchOrg:     c.getOrgID(organizationID),
	}
	response, err := c.client.Subscriptions.DeleteSubscription(&params, c.auth)
	if err != nil {
		return nil, deleteSubscriptionSwaggerError(err)
	}
	return response.Payload, nil
}

func deleteSubscriptionSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *subscriptions.DeleteSubscriptionBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *subscriptions.DeleteSubscriptionUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *subscriptions.DeleteSubscriptionForbidden:
		return NewErrorForbidden(v.Payload)
	case *subscriptions.DeleteSubscriptionNotFound:
		return NewErrorNotFound(v.Payload)
	case *subscriptions.DeleteSubscriptionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetSubscription gets a subscription by name
func (c *DefaultEventsClient) GetSubscription(ctx context.Context, organizationID string, subscriptionName string) (*v1.Subscription, error) {
	params := subscriptions.GetSubscriptionParams{
		Context:          ctx,
		SubscriptionName: subscriptionName,
		XDispatchOrg:     c.getOrgID(organizationID),
	}
	response, err := c.client.Subscriptions.GetSubscription(&params, c.auth)
	if err != nil {
		return nil, getSubscriptionSwaggerError(err)
	}
	return response.Payload, nil
}

func getSubscriptionSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *subscriptions.GetSubscriptionBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *subscriptions.GetSubscriptionUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *subscriptions.GetSubscriptionForbidden:
		return NewErrorForbidden(v.Payload)
	case *subscriptions.GetSubscriptionNotFound:
		return NewErrorNotFound(v.Payload)
	case *subscriptions.GetSubscriptionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListSubscriptions lists all subscriptions
func (c *DefaultEventsClient) ListSubscriptions(ctx context.Context, organizationID string) ([]v1.Subscription, error) {
	params := subscriptions.GetSubscriptionsParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Subscriptions.GetSubscriptions(&params, c.auth)
	if err != nil {
		return nil, listSubscriptionsSwaggerError(err)
	}
	subscriptions := []v1.Subscription{}
	for _, f := range response.Payload {
		subscriptions = append(subscriptions, *f)
	}
	return subscriptions, nil
}

func listSubscriptionsSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *subscriptions.GetSubscriptionsUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *subscriptions.GetSubscriptionsForbidden:
		return NewErrorForbidden(v.Payload)
	case *subscriptions.GetSubscriptionsDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateSubscription updates a specific subscription
func (c *DefaultEventsClient) UpdateSubscription(ctx context.Context, organizationID string, subscription *v1.Subscription) (*v1.Subscription, error) {
	params := subscriptions.UpdateSubscriptionParams{
		Context:          ctx,
		Body:             subscription,
		SubscriptionName: *subscription.Name,
		XDispatchOrg:     c.getOrgID(organizationID),
	}
	response, err := c.client.Subscriptions.UpdateSubscription(&params, c.auth)
	if err != nil {
		return nil, updateSubscriptionSwaggerError(err)
	}
	return response.Payload, nil
}

func updateSubscriptionSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *subscriptions.UpdateSubscriptionBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *subscriptions.UpdateSubscriptionUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *subscriptions.UpdateSubscriptionForbidden:
		return NewErrorForbidden(v.Payload)
	case *subscriptions.UpdateSubscriptionNotFound:
		return NewErrorNotFound(v.Payload)
	case *subscriptions.UpdateSubscriptionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// CreateEventDriver creates and adds a new event driver
func (c *DefaultEventsClient) CreateEventDriver(ctx context.Context, organizationID string, driver *v1.EventDriver) (*v1.EventDriver, error) {
	params := drivers.AddDriverParams{
		Context:      ctx,
		Body:         driver,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.AddDriver(&params, c.auth)
	if err != nil {
		return nil, createDriverSwaggerError(err)
	}
	return response.Payload, nil
}

func createDriverSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.AddDriverBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *drivers.AddDriverUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.AddDriverForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.AddDriverConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *drivers.AddDriverDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteEventDriver deletes a driver
func (c *DefaultEventsClient) DeleteEventDriver(ctx context.Context, organizationID string, driverName string) (*v1.EventDriver, error) {
	params := drivers.DeleteDriverParams{
		Context:      ctx,
		DriverName:   driverName,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.DeleteDriver(&params, c.auth)
	if err != nil {
		return nil, deleteDriverSwaggerError(err)
	}
	return response.Payload, nil
}

func deleteDriverSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.DeleteDriverBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *drivers.DeleteDriverUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.DeleteDriverForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.DeleteDriverNotFound:
		return NewErrorNotFound(v.Payload)
	case *drivers.DeleteDriverDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetEventDriver gets a driver by name
func (c *DefaultEventsClient) GetEventDriver(ctx context.Context, organizationID string, driverName string) (*v1.EventDriver, error) {
	params := drivers.GetDriverParams{
		Context:      ctx,
		DriverName:   driverName,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.GetDriver(&params, c.auth)
	if err != nil {
		return nil, getDriverSwaggerError(err)
	}
	return response.Payload, nil
}

func getDriverSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.GetDriverBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *drivers.GetDriverUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.GetDriverForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.GetDriverNotFound:
		return NewErrorNotFound(v.Payload)
	case *drivers.GetDriverDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListEventDrivers lists all drivers
func (c *DefaultEventsClient) ListEventDrivers(ctx context.Context, organizationID string) ([]v1.EventDriver, error) {
	params := drivers.GetDriversParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.GetDrivers(&params, c.auth)
	if err != nil {
		return nil, listDriversSwaggerError(err)
	}
	drivers := []v1.EventDriver{}
	for _, f := range response.Payload {
		drivers = append(drivers, *f)
	}
	return drivers, nil
}

func listDriversSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.GetDriversUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.GetDriversForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.GetDriversDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateEventDriver updates a specific driver
func (c *DefaultEventsClient) UpdateEventDriver(ctx context.Context, organizationID string, driver *v1.EventDriver) (*v1.EventDriver, error) {
	params := drivers.UpdateDriverParams{
		Context:      ctx,
		Body:         driver,
		DriverName:   *driver.Name,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.UpdateDriver(&params, c.auth)
	if err != nil {
		return nil, updateDriverSwaggerError(err)
	}
	return response.Payload, nil
}

func updateDriverSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.UpdateDriverBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *drivers.UpdateDriverUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.UpdateDriverForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.UpdateDriverNotFound:
		return NewErrorNotFound(v.Payload)
	case *drivers.UpdateDriverDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// CreateEventDriverType creates and adds a new subscription
func (c *DefaultEventsClient) CreateEventDriverType(ctx context.Context, organizationID string, driverType *v1.EventDriverType) (*v1.EventDriverType, error) {
	params := drivers.AddDriverTypeParams{
		Context:      ctx,
		Body:         driverType,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.AddDriverType(&params, c.auth)
	if err != nil {
		return nil, createDriverTypeSwaggerError(err)
	}
	return response.Payload, nil
}

func createDriverTypeSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.AddDriverTypeBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *drivers.AddDriverTypeUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.AddDriverTypeForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.AddDriverTypeConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *drivers.AddDriverTypeDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteEventDriverType deletes a driver
func (c *DefaultEventsClient) DeleteEventDriverType(ctx context.Context, organizationID string, driverTypeName string) (*v1.EventDriverType, error) {
	params := drivers.DeleteDriverTypeParams{
		Context:        ctx,
		DriverTypeName: driverTypeName,
		XDispatchOrg:   c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.DeleteDriverType(&params, c.auth)
	if err != nil {
		return nil, deleteDriverTypeSwaggerError(err)
	}
	return response.Payload, nil
}

func deleteDriverTypeSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.DeleteDriverTypeBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *drivers.DeleteDriverTypeUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.DeleteDriverTypeForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.DeleteDriverTypeNotFound:
		return NewErrorNotFound(v.Payload)
	case *drivers.DeleteDriverTypeDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetEventDriverType gets a driver by name
func (c *DefaultEventsClient) GetEventDriverType(ctx context.Context, organizationID string, driverTypeName string) (*v1.EventDriverType, error) {
	params := drivers.GetDriverTypeParams{
		Context:        ctx,
		DriverTypeName: driverTypeName,
		XDispatchOrg:   c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.GetDriverType(&params, c.auth)
	if err != nil {
		return nil, getDriverTypeSwaggerError(err)
	}
	return response.Payload, nil
}

func getDriverTypeSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.GetDriverTypeBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *drivers.GetDriverTypeUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.GetDriverTypeForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.GetDriverTypeNotFound:
		return NewErrorNotFound(v.Payload)
	case *drivers.GetDriverTypeDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListEventDriverTypes lists all drivers
func (c *DefaultEventsClient) ListEventDriverTypes(ctx context.Context, organizationID string) ([]v1.EventDriverType, error) {
	params := drivers.GetDriverTypesParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.GetDriverTypes(&params, c.auth)
	if err != nil {
		return nil, listDriverTypesSwaggerError(err)
	}
	drivers := []v1.EventDriverType{}
	for _, f := range response.Payload {
		drivers = append(drivers, *f)
	}
	return drivers, nil
}

func listDriverTypesSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.GetDriverTypesUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.GetDriverTypesForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.GetDriverTypesDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateEventDriverType updates a specific driver
func (c *DefaultEventsClient) UpdateEventDriverType(ctx context.Context, organizationID string, driverType *v1.EventDriverType) (*v1.EventDriverType, error) {
	params := drivers.UpdateDriverTypeParams{
		Context:        ctx,
		Body:           driverType,
		DriverTypeName: *driverType.Name,
		XDispatchOrg:   c.getOrgID(organizationID),
	}
	response, err := c.client.Drivers.UpdateDriverType(&params, c.auth)
	if err != nil {
		return nil, updateDriverTypeSwaggerError(err)
	}
	return response.Payload, nil
}

func updateDriverTypeSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *drivers.UpdateDriverTypeBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *drivers.UpdateDriverTypeUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *drivers.UpdateDriverTypeForbidden:
		return NewErrorForbidden(v.Payload)
	case *drivers.UpdateDriverTypeNotFound:
		return NewErrorNotFound(v.Payload)
	case *drivers.UpdateDriverTypeDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}
