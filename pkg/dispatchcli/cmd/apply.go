///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/application-manager/gen/client/application"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	pkgUtils "github.com/vmware/dispatch/pkg/utils"
)

var (
	applyLong = i18n.T(`Update a resource. See subcommands for resources that can be updated.`)

	// TODO: Add examples
	applyExample = i18n.T(``)
)

// NewCmdApply updates command responsible for secret updates.
func NewCmdApply(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apply",
		Short:   i18n.T("Apply resources."),
		Long:    applyLong,
		Example: applyExample,
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
			iamClient := identityManagerClient()

			applyMap := map[string]ModelAction{
				pkgUtils.APIKind:            CallApplyAPI(apiClient),
				pkgUtils.ApplicationKind:    CallApplyApplication,
				pkgUtils.BaseImageKind:      CallApplyBaseImage(imgClient),
				pkgUtils.DriverKind:         CallApplyDriver(eventClient),
				pkgUtils.DriverTypeKind:     CallApplyDriverType(eventClient),
				pkgUtils.FunctionKind:       CallApplyFunction(fnClient),
				pkgUtils.ImageKind:          CallApplyImage(imgClient),
				pkgUtils.SecretKind:         CallApplySecret(secClient),
				pkgUtils.SubscriptionKind:   CallApplySubscription(eventClient),
				pkgUtils.PolicyKind:         CallApplyPolicy(iamClient),
				pkgUtils.ServiceAccountKind: CallApplyServiceAccount(iamClient),
				pkgUtils.OrganizationKind:   CallApplyOrganization(iamClient),
			}

			err := importFile(out, errOut, cmd, args, applyMap, "Applied")
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Working directory relative paths are based on")

	return cmd
}

// CallApplyAPI makes the backend service call to update/create an api
func CallApplyAPI(c client.APIsClient) ModelAction {
	return func(input interface{}) error {
		apiBody := input.(*v1.API)

		_, err := c.UpdateAPI(context.TODO(), "", apiBody)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateAPI(context.TODO(), "", apiBody)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	}
}

// CallApplyApplication makes the API call to update/create an application
func CallApplyApplication(input interface{}) error {
	client := applicationManagerClient()
	body := input.(*v1.Application)
	applicationBody := input.(*v1.Application)

	params := application.NewUpdateAppParams()
	params.Application = *applicationBody.Name
	params.Body = applicationBody
	params.XDispatchOrg = getOrgFromConfig()
	_, err := client.Application.UpdateApp(params, GetAuthInfoWriter())
	if err != nil {
		if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
			params := &application.AddAppParams{
				Body:         body,
				Context:      context.Background(),
				XDispatchOrg: getOrgFromConfig(),
			}

			created, err := client.Application.AddApp(params, GetAuthInfoWriter())
			if err != nil {
				return err
			}
			*body = *created.Payload
			return nil
		}
		return err
	}

	return err
}

// CallApplyBaseImage update/create a base image
func CallApplyBaseImage(c client.ImagesClient) ModelAction {
	return func(input interface{}) error {
		baseImage := input.(*v1.BaseImage)
		_, err := c.UpdateBaseImage(context.TODO(), "", baseImage)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateBaseImage(context.TODO(), dispatchConfig.Organization, baseImage)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	}
}

// CallApplyDriver makes the API call to update/create an event driver
func CallApplyDriver(c client.EventsClient) ModelAction {
	return func(input interface{}) error {
		eventDriver := input.(*v1.EventDriver)

		_, err := c.UpdateEventDriver(context.TODO(), "", eventDriver)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateEventDriver(context.TODO(), "", eventDriver)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	}
}

// CallApplyDriverType makes the API call to update/create a driver type
func CallApplyDriverType(c client.EventsClient) ModelAction {
	return func(input interface{}) error {
		driverType := input.(*v1.EventDriverType)

		_, err := c.UpdateEventDriverType(context.TODO(), "", driverType)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateEventDriverType(context.TODO(), "", driverType)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	}
}

// CallApplyFunction makes the API call to update/create a function
func CallApplyFunction(c client.FunctionsClient) ModelAction {
	return func(input interface{}) error {
		function := input.(*v1.Function)

		_, err := c.UpdateFunction(context.TODO(), "", function)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateFunction(context.TODO(), "", function)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	}
}

// CallApplyImage makes the service call to update/create an image.
func CallApplyImage(c client.ImagesClient) ModelAction {
	return func(input interface{}) error {
		img := input.(*v1.Image)
		_, err := c.UpdateImage(context.TODO(), "", img)

		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateImage(context.TODO(), dispatchConfig.Organization, img)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	}
}

// CallApplyPolicy updates/create a policy
func CallApplyPolicy(c client.IdentityClient) ModelAction {
	return func(p interface{}) error {
		policyModel := p.(*v1.Policy)

		_, err := c.UpdatePolicy(context.TODO(), "", policyModel)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreatePolicy(context.TODO(), "", policyModel)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	}
}

// CallApplyServiceAccount updates/create a serviceaccount
func CallApplyServiceAccount(c client.IdentityClient) ModelAction {
	return func(p interface{}) error {
		serviceaccountModel := p.(*v1.ServiceAccount)

		_, err := c.UpdateServiceAccount(context.TODO(), "", serviceaccountModel)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateServiceAccount(context.TODO(), "", serviceaccountModel)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	}
}

// CallApplyOrganization updates/create an organization
func CallApplyOrganization(c client.IdentityClient) ModelAction {
	return func(p interface{}) error {
		orgModel := p.(*v1.Organization)

		_, err := c.UpdateOrganization(context.TODO(), "", orgModel)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateOrganization(context.TODO(), "", orgModel)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	}
}

// CallApplySecret makes the API call to update/create a secret
func CallApplySecret(c client.SecretsClient) ModelAction {
	return func(input interface{}) error {
		secretModel := input.(*v1.Secret)

		_, err := c.UpdateSecret(context.TODO(), "", secretModel)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateSecret(context.TODO(), dispatchConfig.Organization, secretModel)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	}
}

// CallApplySubscription makes the API call to update/create a subscription
func CallApplySubscription(c client.EventsClient) ModelAction {
	return func(input interface{}) error {
		subscription := input.(*v1.Subscription)

		_, err := c.UpdateSubscription(context.TODO(), "", subscription)
		if err != nil {
			if strings.HasPrefix(fmt.Sprint(err), "[Code: 404] ") {
				_, err := c.CreateSubscription(context.TODO(), "", subscription)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return nil
	}
}
