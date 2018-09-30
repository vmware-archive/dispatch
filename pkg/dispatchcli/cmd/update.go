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
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
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

			fnClient := functionsClient()
			imgClient := imagesClient()
			baseImgClient := baseImagesClient()
			eventClient := eventManagerClient()
			endptClient := endpointsClient()
			secClient := secretsClient()
			iamClient := identityManagerClient()

			updateMap := map[string]ModelAction{
				v1.EndpointKind:       CallUpdateEndpoint(endptClient),
				v1.BaseImageKind:      CallUpdateBaseImage(baseImgClient),
				v1.DriverKind:         CallUpdateDriver(eventClient),
				v1.DriverTypeKind:     CallUpdateDriverType(eventClient),
				v1.FunctionKind:       CallUpdateFunction(fnClient),
				v1.ImageKind:          CallUpdateImage(imgClient),
				v1.SecretKind:         CallUpdateSecret(secClient),
				v1.SubscriptionKind:   CallUpdateSubscription(eventClient),
				v1.PolicyKind:         CallUpdatePolicy(iamClient),
				v1.ServiceAccountKind: CallUpdateServiceAccount(iamClient),
				v1.OrganizationKind:   CallUpdateOrganization(iamClient),
			}

			err := importFile(out, errOut, cmd, args, updateMap, "Updated")
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Working directory relative paths are based on")

	return cmd
}

// CallUpdateEndpoint makes the backend service call to update an api
func CallUpdateEndpoint(c client.EndpointsClient) ModelAction {
	return func(input interface{}) error {
		model := input.(*v1.Endpoint)

		_, err := c.UpdateEndpoint(context.TODO(), "", model)
		if err != nil {
			return err
		}

		return nil
	}
}

// CallUpdateBaseImage updates a base image
func CallUpdateBaseImage(c client.BaseImagesClient) ModelAction {
	return func(input interface{}) error {
		baseImage := input.(*v1.BaseImage)
		_, err := c.UpdateBaseImage(context.TODO(), "", baseImage)
		if err != nil {
			return err
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
			return err
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
			return err
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
			return err
		}

		return nil
	}
}

// CallUpdatePolicy updates a policy
func CallUpdatePolicy(c client.IdentityClient) ModelAction {
	return func(p interface{}) error {

		policyModel := p.(*v1.Policy)

		_, err := c.UpdatePolicy(context.TODO(), "", policyModel)
		if err != nil {
			return nil
		}

		return nil
	}
}

// CallUpdateServiceAccount updates a serviceaccount
func CallUpdateServiceAccount(c client.IdentityClient) ModelAction {
	return func(p interface{}) error {

		serviceaccountModel := p.(*v1.ServiceAccount)

		_, err := c.UpdateServiceAccount(context.TODO(), "", serviceaccountModel)
		if err != nil {
			return err
		}
		return nil
	}
}

// CallUpdateOrganization updates an organization
func CallUpdateOrganization(c client.IdentityClient) ModelAction {
	return func(p interface{}) error {

		orgModel := p.(*v1.Organization)

		_, err := c.UpdateOrganization(context.TODO(), "", orgModel)
		if err != nil {
			return err
		}
		return nil
	}
}

// CallUpdateSecret makes the API call to update a secret
func CallUpdateSecret(c client.SecretsClient) ModelAction {
	return func(input interface{}) error {
		secretModel := input.(*v1.Secret)

		_, err := c.UpdateSecret(context.TODO(), "", secretModel)

		if err != nil {
			return err
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
			return err
		}

		return nil
	}
}
