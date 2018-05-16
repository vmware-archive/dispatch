///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api-manager/gen/client/endpoint"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/application-manager/gen/client/application"
	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/subscriptions"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/policy"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/serviceaccount"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/base_image"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	"github.com/vmware/dispatch/pkg/secret-store/gen/client/secret"
	pkgUtils "github.com/vmware/dispatch/pkg/utils"
)

var (
	updateLong = i18n.T(`Update a resource. See subcommands for resources that can be updated.`)

	// TODO: Add examples
	updateExample = i18n.T(``)
)

// NewCmdUpdate updates command responsible for secret updates.
func NewCmdUpdate(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   i18n.T("Update resources."),
		Long:    updateLong,
		Example: updateExample,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" {
				runHelp(cmd, args)
				return
			}

			updateMap := map[string]modelAction{
				pkgUtils.APIKind:            CallUpdateAPI,
				pkgUtils.ApplicationKind:    CallUpdateApplication,
				pkgUtils.BaseImageKind:      CallUpdateBaseImage,
				pkgUtils.DriverKind:         CallUpdateDriver,
				pkgUtils.DriverTypeKind:     CallUpdateDriverType,
				pkgUtils.FunctionKind:       CallUpdateFunction,
				pkgUtils.ImageKind:          CallUpdateImage,
				pkgUtils.SecretKind:         CallUpdateSecret,
				pkgUtils.SubscriptionKind:   CallUpdateSubscription,
				pkgUtils.PolicyKind:         CallUpdatePolicy,
				pkgUtils.ServiceAccountKind: CallUpdateServiceAccount,
			}

			err := importFile(out, errOut, cmd, args, updateMap)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Working directory relative paths are based on")

	return cmd
}

// CallUpdateAPI makes the backend service call to update an api
func CallUpdateAPI(input interface{}) error {
	apiBody := input.(*v1.API)

	params := endpoint.NewUpdateAPIParams()
	params.API = *apiBody.Name
	params.Body = apiBody

	_, err := apiManagerClient().Endpoint.UpdateAPI(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// CallUpdateApplication makes the API call to update an application
func CallUpdateApplication(input interface{}) error {
	client := applicationManagerClient()
	applicationBody := input.(*v1.Application)

	params := application.NewUpdateAppParams()
	params.Application = *applicationBody.Name
	params.Body = applicationBody
	_, err := client.Application.UpdateApp(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return err
}

// CallUpdateBaseImage updates a base image
func CallUpdateBaseImage(input interface{}) error {
	baseImage := input.(*v1.BaseImage)
	params := base_image.NewUpdateBaseImageByNameParams()
	params.BaseImageName = *baseImage.Name
	params.Body = baseImage
	_, err := imageManagerClient().BaseImage.UpdateBaseImageByName(params, GetAuthInfoWriter())

	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// CallUpdateDriver makes the API call to update an event driver
func CallUpdateDriver(input interface{}) error {
	eventDriver := input.(*v1.EventDriver)
	params := drivers.NewUpdateDriverParams()
	params.DriverName = *eventDriver.Name
	params.Body = eventDriver
	_, err := eventManagerClient().Drivers.UpdateDriver(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// CallUpdateDriverType makes the API call to update a driver type
func CallUpdateDriverType(input interface{}) error {
	driverType := input.(*v1.EventDriverType)
	params := drivers.NewUpdateDriverTypeParams()
	params.DriverTypeName = *driverType.Name
	params.Body = driverType
	_, err := eventManagerClient().Drivers.UpdateDriverType(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// CallUpdateImage makes the service call to update an image.
func CallUpdateImage(input interface{}) error {
	img := input.(*v1.Image)
	params := image.NewUpdateImageByNameParams()
	params.ImageName = *img.Name
	params.Body = img
	_, err := imageManagerClient().Image.UpdateImageByName(params, GetAuthInfoWriter())

	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// CallUpdatePolicy updates a policy
func CallUpdatePolicy(p interface{}) error {

	policyModel := p.(*v1.Policy)

	params := &policy.UpdatePolicyParams{
		PolicyName: *policyModel.Name,
		Body:       policyModel,
		Context:    context.Background(),
	}

	_, err := identityManagerClient().Policy.UpdatePolicy(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// CallUpdateServiceAccount updates a serviceaccount
func CallUpdateServiceAccount(p interface{}) error {

	serviceaccountModel := p.(*v1.ServiceAccount)

	params := &serviceaccount.UpdateServiceAccountParams{
		ServiceAccountName: *serviceaccountModel.Name,
		Body:               serviceaccountModel,
		Context:            context.Background(),
	}

	_, err := identityManagerClient().Serviceaccount.UpdateServiceAccount(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// CallUpdateSecret makes the API call to update a secret
func CallUpdateSecret(input interface{}) error {
	client := secretStoreClient()
	secretBody := input.(*v1.Secret)

	params := secret.NewUpdateSecretParams()
	params.Secret = secretBody
	params.SecretName = *secretBody.Name
	params.Tags = []string{}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	_, err := client.Secret.UpdateSecret(params, GetAuthInfoWriter())

	if err != nil {
		return formatAPIError(err, params)
	}

	return err
}

// CallUpdateSubscription makes the API call to update a subscription
func CallUpdateSubscription(input interface{}) error {
	subscription := input.(*v1.Subscription)
	params := subscriptions.NewUpdateSubscriptionParams()
	params.SubscriptionName = *subscription.Name
	params.Body = subscription
	_, err := eventManagerClient().Subscriptions.UpdateSubscription(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}
