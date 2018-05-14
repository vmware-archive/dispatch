///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/application-manager/gen/client/application"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	createApplicationLong = i18n.T(
		`Create dispatch application.

Note:
  Import your own tls certificates if you want to use your own domain name with HTTPS secure connection
		`)
	// TODO: add examples
	createApplicationExample = i18n.T(``)
)

// NewCmdCreateApplication creates command responsible for dispatch application creation.
func NewCmdCreateApplication(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application NAME",
		Short:   i18n.T("Create application"),
		Long:    createApplicationLong,
		Example: createApplicationExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"app"},
		Run: func(cmd *cobra.Command, args []string) {
			err := createApplication(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

// CallCreateApplication makes the API call to create an application
func CallCreateApplication(i interface{}) error {
	client := applicationManagerClient()
	body := i.(*v1.Application)
	params := &application.AddAppParams{
		Body:    body,
		Context: context.Background(),
	}

	created, err := client.Application.AddApp(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*body = *created.Payload
	return nil
}

func createApplication(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	body := &v1.Application{
		Name: swag.String(args[0]),
		Tags: []*v1.Tag{},
	}
	if cmdFlagApplication != "" {
		body.Tags = append(body.Tags, &v1.Tag{Key: "Application", Value: cmdFlagApplication})
	}
	err := CallCreateApplication(body)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*body)
	}
	fmt.Fprintf(out, "created application: %s\n", *body.Name)
	return nil
}
