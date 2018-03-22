///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	models "github.com/vmware/dispatch/pkg/image-manager/gen/models"
)

var (
	updateImageLong    = "dispatch update image my-image"
	updateImageExample = ""
)

// CallUpdateImage makes the service call to update an image.
func CallUpdateImage(input interface{}) error {
	img := input.(*models.Image)
	params := image.NewUpdateImageByNameParams()
	params.ImageName = *img.Name
	params.Body = img
	_, err := imageManagerClient().Image.UpdateImageByName(params, GetAuthInfoWriter())

	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// NewCmdUpdateImage creates command responsible for updating an image.
func NewCmdUpdateImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "image IMAGE_NAME",
		Short:   i18n.T("Update an image"),
		Long:    updateImageLong,
		Example: updateImageExample,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updateImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	return cmd
}

func updateImage(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {
	//TODO: implement file writer update mechanism
	return nil
}
