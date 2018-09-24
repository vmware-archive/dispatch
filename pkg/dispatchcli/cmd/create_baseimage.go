///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	createBaseImageLong = i18n.T(`Create base image.`)

	// TODO: add examples
	createBaseImageExample = i18n.T(``)
	public                 = false
	language               = i18n.T(``)
)

// NewCmdCreateBaseImage creates command responsible for base image creation.
func NewCmdCreateBaseImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "base-image IMAGE_NAME IMAGE_URL [--language LANGUAGE]",
		Short:   i18n.T("Create base image"),
		Long:    createBaseImageLong,
		Example: createBaseImageExample,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			c := imageManagerClient()
			err := createBaseImage(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVar(&language, "language", "", "Specify the runtime language for the image")
	return cmd
}

// CallCreateBaseImage makes the API call to create a base image
func CallCreateBaseImage(c client.ImagesClient) ModelAction {
	return func(bi interface{}) error {
		baseImage := bi.(*v1.BaseImage)

		created, err := c.CreateBaseImage(context.TODO(), dispatchConfig.Organization, baseImage)
		if err != nil {
			return err
		}
		*baseImage = *created
		return nil
	}
}

func createBaseImage(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.ImagesClient) error {
	baseImage := &v1.BaseImage{
		Meta: v1.Meta{
			Name: args[0],
		},
		DockerURL: &args[1],
		Language:  swag.String(language),
	}
	err := CallCreateBaseImage(c)(baseImage)
	if err != nil {
		return err
	}
	if w, err := formatOutput(out, false, baseImage); w {
		return err
	}
	fmt.Fprintf(out, "Created base image: %s\n", baseImage.Name)
	return nil
}
