///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/utils"
)

var (
	applyLong = i18n.T(`Update a resource. See subcommands for resources that can be updated.`)

	// TODO: Add examples
	applyExample = i18n.T(``)
)

var applyMap map[string]ModelAction

func initApplyMap() {
	fnClient := functionManagerClient()
	imgClient := imageManagerClient()
	eventClient := eventManagerClient()
	apiClient := apiManagerClient()
	secClient := secretStoreClient()
	iamClient := identityManagerClient()

	applyMap = map[string]ModelAction{
		utils.ImageKind:          CallApplyImage(imgClient),
		utils.BaseImageKind:      CallApplyBaseImage(imgClient),
		utils.FunctionKind:       CallApplyFunction(fnClient),
		utils.SecretKind:         CallApplySecret(secClient),
		utils.PolicyKind:         CallApplyPolicy(iamClient),
		utils.ServiceAccountKind: CallApplyServiceAccount(iamClient),
		utils.DriverTypeKind:     CallApplyDriverType(eventClient),
		utils.DriverKind:         CallApplyDriver(eventClient),
		utils.SubscriptionKind:   CallApplySubscription(eventClient),
		utils.APIKind:            CallApplyAPI(apiClient),
		utils.OrganizationKind:   CallApplyOrganization(iamClient),
	}
}

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

			initApplyMap()
			if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
				isURL = true
				baseURL = file[:strings.LastIndex(file, "/")+1]
				err := importFileWithURL(out, errOut, cmd, args, applyMap, "Applied")
				CheckErr(err)
			} else {
				err := importFile(out, errOut, cmd, args, applyMap, "Applied")
				CheckErr(err)
			}
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file or an URL")
	cmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Working directory relative paths are based on")

	cmd.AddCommand(NewCmdApplySeedImages(out, errOut))
	return cmd
}

// NewCmdApplySeedImages apply command responsible for apply of seed images and base-images.
func NewCmdApplySeedImages(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "seed-images",
		Short:   i18n.T("Apply seed base-images and images"),
		Long:    i18n.T(`Apply base-images and images to quick-start with the current version of Dispatch`),
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"seed"},
		Run: func(cmd *cobra.Command, args []string) {
			err := applySeedImages(out)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&outputFile, "output-file", "O", "", "seed images YAML gets written to this file (nothing gets created)")
	return cmd
}

func applySeedImages(out io.Writer) error {
	if imagesB64 == "" {
		return errors.New("embedded images YAML is empty")
	}

	sr := strings.NewReader(imagesB64)
	br := base64.NewDecoder(base64.StdEncoding, sr)
	gr, err := gzip.NewReader(br)
	if err != nil {
		return errors.Wrap(err, "error creating a gzip reader for embedded images YAML")
	}
	bs := &bytes.Buffer{}
	_, err = bs.ReadFrom(gr)
	if err != nil {
		return errors.Wrap(err, "error reading embedded images YAML: error reading from gzip reader")
	}

	if outputFile != "" {
		err := ioutil.WriteFile(outputFile, bs.Bytes(), 0644)
		return errors.Wrapf(err, "error writing images YAML to '%s'", outputFile)
	}

	initApplyMap()
	return importBytes(out, bs.Bytes(), applyMap, "Applied")
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
