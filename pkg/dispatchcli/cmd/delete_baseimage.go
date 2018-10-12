///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	deleteBaseImagesLong = i18n.T(`Delete base images.`)

	// TODO: add examples
	deleteBaseImagesExample = i18n.T(``)
)

// NewCmdDeleteBaseImage creates command responsible for deleting base images.
func NewCmdDeleteBaseImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "base-image IMAGE_NAME",
		Short:   i18n.T("Delete base images"),
		Long:    deleteBaseImagesLong,
		Example: deleteBaseImagesExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"base-images", "baseimage", "baseimages"},
		Run: func(cmd *cobra.Command, args []string) {
			c := baseImagesClient()
			err := deleteBaseImage(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	return cmd
}

// CallDeleteBaseImage makes the API call to create an image
func CallDeleteBaseImage(c client.BaseImagesClient) ModelAction {
	return func(i interface{}) error {
		baseImageModel := i.(*v1.BaseImage)
		_, err := c.DeleteBaseImage(context.TODO(), dispatchConfig.Organization, baseImageModel.Name)
		if err != nil {
			return err
		}
		return nil
	}
}

func deleteBaseImage(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.BaseImagesClient) error {
	baseImageModel := v1.BaseImage{
		Meta: v1.Meta{
			Name: args[0],
		},
	}
	err := CallDeleteBaseImage(c)(&baseImageModel)
	if err != nil {
		return err
	}
	return formatDeleteBaseImageOutput(out, false, []*v1.BaseImage{&baseImageModel})
}

func formatDeleteBaseImageOutput(out io.Writer, list bool, images []*v1.BaseImage) error {
	if w, err := formatOutput(out, list, images); w {
		return err
	}
	for _, i := range images {
		_, err := fmt.Fprintf(out, "Deleted base image: %s\n", i.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
