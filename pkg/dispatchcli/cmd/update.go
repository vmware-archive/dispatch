///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/application-manager/gen/client/application"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/policy"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/serviceaccount"
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

			fnClient := functionManagerClient()
			imgClient := imageManagerClient()
			eventClient := eventManagerClient()
			apiClient := apiManagerClient()
			secClient := secretStoreClient()

			updateMap := map[string]ModelAction{
				pkgUtils.APIKind:            CallUpdateAPI(apiClient),
				pkgUtils.ApplicationKind:    CallUpdateApplication,
				pkgUtils.BaseImageKind:      CallUpdateBaseImage(imgClient),
				pkgUtils.DriverKind:         CallUpdateDriver(eventClient),
				pkgUtils.DriverTypeKind:     CallUpdateDriverType(eventClient),
				pkgUtils.FunctionKind:       CallUpdateFunction(fnClient),
				pkgUtils.ImageKind:          CallUpdateImage(imgClient),
				pkgUtils.SecretKind:         CallUpdateSecret(secClient),
				pkgUtils.SubscriptionKind:   CallUpdateSubscription(eventClient),
				pkgUtils.PolicyKind:         CallUpdatePolicy,
				pkgUtils.ServiceAccountKind: CallUpdateServiceAccount,
			}

			err := importFile(out, errOut, cmd, args, updateMap, "Updated")
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Working directory relative paths are based on")

	return cmd
}

// CallUpdateAPI makes the backend service call to update an api
func CallUpdateAPI(c client.APIsClient) ModelAction {
	return func(input interface{}) error {
		apiBody := input.(*v1.API)

		_, err := c.UpdateAPI(context.TODO(), "", apiBody)
		if err != nil {
			return formatAPIError(err, apiBody)
		}

		return nil
	}
}

// CallUpdateApplication makes the API call to update an application
func CallUpdateApplication(input interface{}) error {
	client := applicationManagerClient()
	applicationBody := input.(*v1.Application)

	params := application.NewUpdateAppParams()
	params.Application = *applicationBody.Name
	params.Body = applicationBody
	params.XDispatchOrg = getOrganization()
	_, err := client.Application.UpdateApp(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return err
}

// CallUpdateBaseImage updates a base image
func CallUpdateBaseImage(c client.ImagesClient) ModelAction {
	return func(input interface{}) error {
		baseImage := input.(*v1.BaseImage)
		_, err := c.UpdateBaseImage(context.TODO(), "", baseImage)
		if err != nil {
			return formatAPIError(err, *baseImage.Name)
		}

		return nil
	}
}

// CallUpdateDriver makes the API call to update an event driver
func CallUpdateDriver(c client.EventsClient) ModelAction {
	return func(input interface{}) error {
		eventDriver := input.(*v1.EventDriver)

		_, err := c.UpdateEventDriver(context.TODO(), "", eventDriver)
		if err != nil {
			return formatAPIError(err, eventDriver)
		}

		return nil
	}
}

// CallUpdateDriverType makes the API call to update a driver type
func CallUpdateDriverType(c client.EventsClient) ModelAction {
	return func(input interface{}) error {
		driverType := input.(*v1.EventDriverType)

		_, err := c.UpdateEventDriverType(context.TODO(), "", driverType)
		if err != nil {
			return formatAPIError(err, driverType)
		}

		return nil
	}
}

// CallUpdateImage makes the service call to update an image.
func CallUpdateImage(c client.ImagesClient) ModelAction {
	return func(input interface{}) error {
		img := input.(*v1.Image)
		_, err := c.UpdateImage(context.TODO(), "", img)

		if err != nil {
			return formatAPIError(err, *img.Name)
		}

		return nil
	}
}

// CallUpdatePolicy updates a policy
func CallUpdatePolicy(p interface{}) error {

	policyModel := p.(*v1.Policy)

	params := &policy.UpdatePolicyParams{
		PolicyName:   *policyModel.Name,
		Body:         policyModel,
		Context:      context.Background(),
		XDispatchOrg: getOrganization(),
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
		XDispatchOrg:       getOrganization(),
	}

	_, err := identityManagerClient().Serviceaccount.UpdateServiceAccount(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// CallUpdateSecret makes the API call to update a secret
func CallUpdateSecret(c client.SecretsClient) ModelAction {
	return func(input interface{}) error {
		secretModel := input.(*v1.Secret)

		_, err := c.UpdateSecret(context.TODO(), "", secretModel)

		if err != nil {
			return formatAPIError(err, secretModel.Name)
		}
		return err
	}
}

// CallUpdateSubscription makes the API call to update a subscription
func CallUpdateSubscription(c client.EventsClient) ModelAction {
	return func(input interface{}) error {
		subscription := input.(*v1.Subscription)

		_, err := c.UpdateSubscription(context.TODO(), "", subscription)
		if err != nil {
			return formatAPIError(err, subscription)
		}

		return nil
	}
}
