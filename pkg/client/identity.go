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

	swaggerclient "github.com/vmware/dispatch/pkg/identity-manager/gen/client"
	swaggerops "github.com/vmware/dispatch/pkg/identity-manager/gen/client/operations"
	swaggerorgs "github.com/vmware/dispatch/pkg/identity-manager/gen/client/organization"
	swaggerpolicy "github.com/vmware/dispatch/pkg/identity-manager/gen/client/policy"
	swaggeraccounts "github.com/vmware/dispatch/pkg/identity-manager/gen/client/serviceaccount"
)

// IdentityClient defines the identity client interface
type IdentityClient interface {
	// Policies
	CreatePolicy(ctx context.Context, organizationID string, policy *v1.Policy) (*v1.Policy, error)
	DeletePolicy(ctx context.Context, organizationID string, policyName string) (*v1.Policy, error)
	UpdatePolicy(ctx context.Context, organizationID string, policy *v1.Policy) (*v1.Policy, error)
	GetPolicy(ctx context.Context, organizationID string, policyName string) (*v1.Policy, error)
	ListPolicies(ctx context.Context, organizationID string) ([]v1.Policy, error)

	// Organizations
	CreateOrganization(ctx context.Context, organizationID string, org *v1.Organization) (*v1.Organization, error)
	DeleteOrganization(ctx context.Context, organizationID string, orgName string) (*v1.Organization, error)
	UpdateOrganization(ctx context.Context, organizationID string, org *v1.Organization) (*v1.Organization, error)
	GetOrganization(ctx context.Context, organizationID string, orgName string) (*v1.Organization, error)
	ListOrganizations(ctx context.Context, organizationID string) ([]v1.Organization, error)

	// Service Accounts
	CreateServiceAccount(ctx context.Context, organizationID string, svcAccount *v1.ServiceAccount) (*v1.ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, organizationID string, svcAccountName string) (*v1.ServiceAccount, error)
	UpdateServiceAccount(ctx context.Context, organizationID string, svcAccount *v1.ServiceAccount) (*v1.ServiceAccount, error)
	GetServiceAccount(ctx context.Context, organizationID string, svcAccountName string) (*v1.ServiceAccount, error)
	ListServiceAccounts(ctx context.Context, organizationID string) ([]v1.ServiceAccount, error)

	// Other operations
	GetVersion(ctx context.Context) (*v1.Version, error)
	Home(ctx context.Context, organizationID string) (*v1.Message, error)
}

// DefaultIdentityClient defines the default client for events API
type DefaultIdentityClient struct {
	baseClient

	client *swaggerclient.IdentityManager
	auth   runtime.ClientAuthInfoWriter
}

// NewIdentityClient is used to create a new subscriptions client
func NewIdentityClient(host string, auth runtime.ClientAuthInfoWriter, organizationID string) *DefaultIdentityClient {
	transport := DefaultHTTPClient(host, swaggerclient.DefaultBasePath)
	return &DefaultIdentityClient{
		baseClient: baseClient{
			organizationID: organizationID,
		},
		client: swaggerclient.New(transport, strfmt.Default),
		auth:   auth,
	}
}

// CreatePolicy creates new policy
func (c *DefaultIdentityClient) CreatePolicy(ctx context.Context, organizationID string, policy *v1.Policy) (*v1.Policy, error) {
	params := swaggerpolicy.AddPolicyParams{
		Body:         policy,
		XDispatchOrg: c.getOrgID(organizationID),
		Context:      ctx,
	}
	response, err := c.client.Policy.AddPolicy(&params, c.auth)
	if err != nil {
		return nil, createPolicySwaggerError(err)
	}
	return response.Payload, nil
}

func createPolicySwaggerError(err error) error {
	switch v := err.(type) {
	case *swaggerpolicy.AddPolicyBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggerpolicy.AddPolicyUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerpolicy.AddPolicyForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerpolicy.AddPolicyConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *swaggerpolicy.AddPolicyDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeletePolicy deletes the policy
func (c *DefaultIdentityClient) DeletePolicy(ctx context.Context, organizationID string, policyName string) (*v1.Policy, error) {
	params := swaggerpolicy.DeletePolicyParams{
		PolicyName:   policyName,
		XDispatchOrg: c.getOrgID(organizationID),
		Context:      ctx,
	}
	response, err := c.client.Policy.DeletePolicy(&params, c.auth)
	if err != nil {
		return nil, deletePolicySwaggerError(err)
	}
	return response.Payload, nil
}

func deletePolicySwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerpolicy.DeletePolicyBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggerpolicy.DeletePolicyUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerpolicy.DeletePolicyForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerpolicy.DeletePolicyNotFound:
		return NewErrorNotFound(v.Payload)
	case *swaggerpolicy.DeletePolicyDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdatePolicy updates the policy
func (c *DefaultIdentityClient) UpdatePolicy(ctx context.Context, organizationID string, policy *v1.Policy) (*v1.Policy, error) {
	params := swaggerpolicy.UpdatePolicyParams{
		PolicyName:   *policy.Name,
		XDispatchOrg: c.getOrgID(organizationID),
		Body:         policy,
		Context:      ctx,
	}
	response, err := c.client.Policy.UpdatePolicy(&params, c.auth)
	if err != nil {
		return nil, updatePolicySwaggerError(err)
	}
	return response.Payload, nil
}

func updatePolicySwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerpolicy.UpdatePolicyBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggerpolicy.UpdatePolicyUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerpolicy.UpdatePolicyForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerpolicy.UpdatePolicyNotFound:
		return NewErrorNotFound(v.Payload)
	case *swaggerpolicy.UpdatePolicyDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetPolicy deletes the policy
func (c *DefaultIdentityClient) GetPolicy(ctx context.Context, organizationID string, policyName string) (*v1.Policy, error) {
	params := swaggerpolicy.GetPolicyParams{
		PolicyName:   policyName,
		XDispatchOrg: c.getOrgID(organizationID),
		Context:      ctx,
	}
	response, err := c.client.Policy.GetPolicy(&params, c.auth)
	if err != nil {
		return nil, getPolicySwaggerError(err)
	}
	return response.Payload, nil
}

func getPolicySwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerpolicy.GetPolicyBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggerpolicy.GetPolicyUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerpolicy.GetPolicyForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerpolicy.GetPolicyNotFound:
		return NewErrorNotFound(v.Payload)
	case *swaggerpolicy.GetPolicyDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListPolicies lists all functions
func (c *DefaultIdentityClient) ListPolicies(ctx context.Context, organizationID string) ([]v1.Policy, error) {
	params := swaggerpolicy.GetPoliciesParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Policy.GetPolicies(&params, c.auth)
	if err != nil {
		return nil, listPoliciesSwaggerError(err)
	}
	policies := []v1.Policy{}
	for _, f := range response.Payload {
		policies = append(policies, *f)
	}
	return policies, nil
}

func listPoliciesSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerpolicy.GetPoliciesUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerpolicy.GetPoliciesForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerpolicy.GetPoliciesDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// CreateOrganization creates new policy
func (c *DefaultIdentityClient) CreateOrganization(ctx context.Context, organizationID string, policy *v1.Organization) (*v1.Organization, error) {
	orgID := c.getOrgID(organizationID)
	params := swaggerorgs.AddOrganizationParams{
		Body:         policy,
		XDispatchOrg: &orgID,
		Context:      ctx,
	}
	response, err := c.client.Organization.AddOrganization(&params, c.auth)
	if err != nil {
		return nil, createOrganizationSwaggerError(err)
	}
	return response.Payload, nil
}

func createOrganizationSwaggerError(err error) error {
	switch v := err.(type) {
	case *swaggerorgs.AddOrganizationBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggerorgs.AddOrganizationUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerorgs.AddOrganizationForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerorgs.AddOrganizationConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *swaggerorgs.AddOrganizationDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteOrganization deletes the policy
func (c *DefaultIdentityClient) DeleteOrganization(ctx context.Context, organizationID string, policyName string) (*v1.Organization, error) {
	params := swaggerorgs.DeleteOrganizationParams{
		OrganizationName: policyName,
		XDispatchOrg:     c.getOrgID(organizationID),
		Context:          ctx,
	}
	response, err := c.client.Organization.DeleteOrganization(&params, c.auth)
	if err != nil {
		return nil, deleteOrganizationSwaggerError(err)
	}
	return response.Payload, nil
}

func deleteOrganizationSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerorgs.DeleteOrganizationBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggerorgs.DeleteOrganizationUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerorgs.DeleteOrganizationForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerorgs.DeleteOrganizationNotFound:
		return NewErrorNotFound(v.Payload)
	case *swaggerorgs.DeleteOrganizationDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateOrganization updates the policy
func (c *DefaultIdentityClient) UpdateOrganization(ctx context.Context, organizationID string, policy *v1.Organization) (*v1.Organization, error) {
	params := swaggerorgs.UpdateOrganizationParams{
		OrganizationName: *policy.Name,
		Body:             policy,
		XDispatchOrg:     c.getOrgID(organizationID),
		Context:          ctx,
	}
	response, err := c.client.Organization.UpdateOrganization(&params, c.auth)
	if err != nil {
		return nil, updateOrganizationSwaggerError(err)
	}
	return response.Payload, nil
}

func updateOrganizationSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerorgs.UpdateOrganizationBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggerorgs.UpdateOrganizationUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerorgs.UpdateOrganizationForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerorgs.UpdateOrganizationNotFound:
		return NewErrorNotFound(v.Payload)
	case *swaggerorgs.UpdateOrganizationDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetOrganization deletes the policy
func (c *DefaultIdentityClient) GetOrganization(ctx context.Context, organizationID string, policyName string) (*v1.Organization, error) {
	params := swaggerorgs.GetOrganizationParams{
		OrganizationName: policyName,
		XDispatchOrg:     c.getOrgID(organizationID),
		Context:          ctx,
	}
	response, err := c.client.Organization.GetOrganization(&params, c.auth)
	if err != nil {
		return nil, getOrganizationSwaggerError(err)
	}
	return response.Payload, nil
}

func getOrganizationSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerorgs.GetOrganizationBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggerorgs.GetOrganizationUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerorgs.GetOrganizationForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerorgs.GetOrganizationNotFound:
		return NewErrorNotFound(v.Payload)
	case *swaggerorgs.GetOrganizationDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListOrganizations lists all functions
func (c *DefaultIdentityClient) ListOrganizations(ctx context.Context, organizationID string) ([]v1.Organization, error) {
	orgID := c.getOrgID(organizationID)
	params := swaggerorgs.GetOrganizationsParams{
		Context:      ctx,
		XDispatchOrg: &orgID,
	}
	response, err := c.client.Organization.GetOrganizations(&params, c.auth)
	if err != nil {
		return nil, listOrganizationsSwaggerError(err)
	}
	policies := []v1.Organization{}
	for _, f := range response.Payload {
		policies = append(policies, *f)
	}
	return policies, nil
}

func listOrganizationsSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerorgs.GetOrganizationsUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerorgs.GetOrganizationsForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerorgs.GetOrganizationsDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// CreateServiceAccount creates new policy
func (c *DefaultIdentityClient) CreateServiceAccount(ctx context.Context, organizationID string, policy *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	params := swaggeraccounts.AddServiceAccountParams{
		Body:         policy,
		XDispatchOrg: c.getOrgID(organizationID),
		Context:      ctx,
	}
	response, err := c.client.Serviceaccount.AddServiceAccount(&params, c.auth)
	if err != nil {
		return nil, createServiceAccountSwaggerError(err)
	}
	return response.Payload, nil
}

func createServiceAccountSwaggerError(err error) error {
	switch v := err.(type) {
	case *swaggeraccounts.AddServiceAccountBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggeraccounts.AddServiceAccountUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggeraccounts.AddServiceAccountForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggeraccounts.AddServiceAccountConflict:
		return NewErrorAlreadyExists(v.Payload)
	case *swaggeraccounts.AddServiceAccountDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// DeleteServiceAccount deletes the policy
func (c *DefaultIdentityClient) DeleteServiceAccount(ctx context.Context, organizationID string, policyName string) (*v1.ServiceAccount, error) {
	params := swaggeraccounts.DeleteServiceAccountParams{
		ServiceAccountName: policyName,
		XDispatchOrg:       c.getOrgID(organizationID),
		Context:            ctx,
	}
	response, err := c.client.Serviceaccount.DeleteServiceAccount(&params, c.auth)
	if err != nil {
		return nil, deleteServiceAccountSwaggerError(err)
	}
	return response.Payload, nil
}

func deleteServiceAccountSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggeraccounts.DeleteServiceAccountBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggeraccounts.DeleteServiceAccountUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggeraccounts.DeleteServiceAccountForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggeraccounts.DeleteServiceAccountNotFound:
		return NewErrorNotFound(v.Payload)
	case *swaggeraccounts.DeleteServiceAccountDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// UpdateServiceAccount updates the policy
func (c *DefaultIdentityClient) UpdateServiceAccount(ctx context.Context, organizationID string, policy *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	params := swaggeraccounts.UpdateServiceAccountParams{
		ServiceAccountName: *policy.Name,
		Body:               policy,
		XDispatchOrg:       c.getOrgID(organizationID),
		Context:            ctx,
	}
	response, err := c.client.Serviceaccount.UpdateServiceAccount(&params, c.auth)
	if err != nil {
		return nil, updateServiceAccountSwaggerError(err)
	}
	return response.Payload, nil
}

func updateServiceAccountSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggeraccounts.UpdateServiceAccountBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggeraccounts.UpdateServiceAccountUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggeraccounts.UpdateServiceAccountForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggeraccounts.UpdateServiceAccountNotFound:
		return NewErrorNotFound(v.Payload)
	case *swaggeraccounts.UpdateServiceAccountDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetServiceAccount deletes the policy
func (c *DefaultIdentityClient) GetServiceAccount(ctx context.Context, organizationID string, policyName string) (*v1.ServiceAccount, error) {
	params := swaggeraccounts.GetServiceAccountParams{
		ServiceAccountName: policyName,
		XDispatchOrg:       c.getOrgID(organizationID),
		Context:            ctx,
	}
	response, err := c.client.Serviceaccount.GetServiceAccount(&params, c.auth)
	if err != nil {
		return nil, getServiceAccountSwaggerError(err)
	}
	return response.Payload, nil
}

func getServiceAccountSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggeraccounts.GetServiceAccountBadRequest:
		return NewErrorBadRequest(v.Payload)
	case *swaggeraccounts.GetServiceAccountUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggeraccounts.GetServiceAccountForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggeraccounts.GetServiceAccountNotFound:
		return NewErrorNotFound(v.Payload)
	case *swaggeraccounts.GetServiceAccountDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// ListServiceAccounts lists all functions
func (c *DefaultIdentityClient) ListServiceAccounts(ctx context.Context, organizationID string) ([]v1.ServiceAccount, error) {
	params := swaggeraccounts.GetServiceAccountsParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Serviceaccount.GetServiceAccounts(&params, c.auth)
	if err != nil {
		return nil, listServiceAccountsSwaggerError(err)
	}
	policies := []v1.ServiceAccount{}
	for _, f := range response.Payload {
		policies = append(policies, *f)
	}
	return policies, nil
}

func listServiceAccountsSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggeraccounts.GetServiceAccountsUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggeraccounts.GetServiceAccountsForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggeraccounts.GetServiceAccountsDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// GetVersion retrievies version from Dispatch
func (c *DefaultIdentityClient) GetVersion(ctx context.Context) (*v1.Version, error) {
	params := swaggerops.GetVersionParams{
		Context: ctx,
	}
	response, err := c.client.Operations.GetVersion(&params)
	if err != nil {
		return nil, getVersionSwaggerError(err)
	}
	return response.Payload, nil
}

func getVersionSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerops.GetVersionDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}

// Home checks availability of Dispatch
func (c *DefaultIdentityClient) Home(ctx context.Context, organizationID string) (*v1.Message, error) {
	params := swaggerops.HomeParams{
		Context:      ctx,
		XDispatchOrg: c.getOrgID(organizationID),
	}
	response, err := c.client.Operations.Home(&params, c.auth)
	if err != nil {
		return nil, homeSwaggerError(err)
	}
	return response.Payload, nil
}

func homeSwaggerError(err error) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *swaggerops.HomeUnauthorized:
		return NewErrorUnauthorized(v.Payload)
	case *swaggerops.HomeForbidden:
		return NewErrorForbidden(v.Payload)
	case *swaggerops.HomeDefault:
		return NewErrorServerUnknownError(v.Payload)
	default:
		// shouldn't happen, but we need to be prepared:
		return fmt.Errorf("unexpected error received from server: %s", err)
	}
}
