///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	createServiceInstanceLong = i18n.T(`Create service instance.`)

	// TODO: add examples
	createServiceInstanceExample = i18n.T(``)
	servicePlan                  = i18n.T(``)
	serviceParameters            = i18n.T(``)
	serviceSecrets               = []string{}
	bindingParamters             = i18n.T(``)
	bindingSecrets               = []string{}
	bindingSecretKey             = i18n.T(``)
)

// CallCreateServiceInstance makes the API call to create a service instance
func CallCreateServiceInstance(c client.ServicesClient) ModelAction {
	return func(s interface{}) error {
		serviceInstanceModel := s.(*v1.ServiceInstance)

		created, err := c.CreateServiceInstance(context.TODO(), dispatchConfig.Organization, serviceInstanceModel)
		if err != nil {
			return err
		}

		*serviceInstanceModel = *created
		return nil
	}
}

// NewCmdCreateServiceInstance creates command responsible for service instance creation.
func NewCmdCreateServiceInstance(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceinstance SERVICE_INSTANCE_NAME SERVICE_CLASS_NAME SERVICE_PLAN_NAME",
		Short:   i18n.T("Create serviceinstance"),
		Long:    createServiceInstanceLong,
		Example: createServiceInstanceExample,
		Args:    cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			c := serviceManagerClient()
			err := createServiceInstance(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringVarP(&serviceParameters, "params", "p", "", "service instance provisioning parameters (JSON)")
	cmd.Flags().StringArrayVarP(&serviceSecrets, "secret", "s", []string{}, "service instance provisioning secrets")
	cmd.Flags().StringVarP(&bindingParamters, "binding-params", "P", "", "service instance binding parameters (JSON)")
	cmd.Flags().StringArrayVarP(&bindingSecrets, "binding-secret", "S", []string{}, "service instance binding secrets")
	cmd.Flags().StringVarP(&bindingSecretKey, "binding-secret-key", "B", "", "service instance binding secret key")
	return cmd
}

func parseParameters(p string) (map[string]interface{}, error) {
	if p != "" {
		m := make(map[string]interface{})
		err := json.Unmarshal([]byte(p), &m)
		if err != nil {
			return nil, fmt.Errorf("parameters must be a JSON map: %s", p)
		}
		return m, nil
	}
	return nil, nil
}

func createServiceInstance(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.ServicesClient) error {
	body := &v1.ServiceInstance{
		Name:             &args[0],
		ServiceClass:     &args[1],
		ServicePlan:      &args[2],
		SecretParameters: serviceSecrets,
		Binding: &v1.ServiceBinding{
			SecretParameters: bindingSecrets,
			BindingSecret:    bindingSecretKey,
		},
	}

	if cmdFlagApplication != "" {
		body.Tags = append(body.Tags, &v1.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

	p, err := parseParameters(serviceParameters)
	if err != nil {
		return err
	}
	body.Parameters = p

	p, err = parseParameters(bindingParamters)
	if err != nil {
		return err
	}
	body.Binding.Parameters = p

	err = CallCreateServiceInstance(c)(body)
	if err != nil {
		return err
	}
	if w, err := formatOutput(out, false, body); w {
		return err
	}
	fmt.Fprintf(out, "Created serviceinstance: %s\n", *body.Name)
	return nil
}
