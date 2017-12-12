///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	baseimage "github.com/vmware/dispatch/pkg/image-manager/gen/client/base_image"
	models "github.com/vmware/dispatch/pkg/image-manager/gen/models"
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
		Aliases: []string{"base-images"},
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteBaseImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

// CallDeleteBaseImage makes the API call to create an image
func CallDeleteBaseImage(i interface{}) error {
	client := imageManagerClient()
	baseImageModel := i.(*models.BaseImage)
	params := &baseimage.DeleteBaseImageByNameParams{
		BaseImageName: *baseImageModel.Name,
		Context:       context.Background(),
	}
	deleted, err := client.BaseImage.DeleteBaseImageByName(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*baseImageModel = *deleted.Payload
	return nil
}

func deleteBaseImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	baseImageModel := models.BaseImage{
		Name: &args[0],
	}
	err := CallDeleteBaseImage(&baseImageModel)
	if err != nil {
		return err
	}
	return formatDeleteBaseImageOutput(out, false, []*models.BaseImage{&baseImageModel})
}

func formatDeleteBaseImageOutput(out io.Writer, list bool, images []*models.BaseImage) error {
	if dispatchConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(images)
		}
		return encoder.Encode(images[0])
	}
	for _, i := range images {
		_, err := fmt.Fprintf(out, "Deleted base image: %s\n", *i.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
