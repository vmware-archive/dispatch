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
	image "github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	models "github.com/vmware/dispatch/pkg/image-manager/gen/models"
)

var (
	deleteImagesLong = i18n.T(`Delete images.`)

	// TODO: add examples
	deleteImagesExample = i18n.T(``)
)

// NewCmdDeleteImage creates command responsible for deleting  images.
func NewCmdDeleteImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "image IMAGE_NAME",
		Short:   i18n.T("Delete images"),
		Long:    deleteImagesLong,
		Example: deleteImagesExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"images"},
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

// CallDeleteImage makes the API call to delete an image
func CallDeleteImage(i interface{}) error {
	client := imageManagerClient()
	imageModel := i.(*models.Image)
	params := &image.DeleteImageByNameParams{
		ImageName: *imageModel.Name,
		Context:   context.Background(),
	}

	deleted, err := client.Image.DeleteImageByName(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*imageModel = *deleted.Payload
	return nil
}

func deleteImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	imageModel := models.Image{
		Name: &args[0],
	}
	err := CallDeleteImage(&imageModel)
	if err != nil {
		return err
	}
	return formatDeleteImageOutput(out, false, []*models.Image{&imageModel})
}

func formatDeleteImageOutput(out io.Writer, list bool, images []*models.Image) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(images)
		}
		return encoder.Encode(images[0])
	}
	for _, i := range images {
		_, err := fmt.Fprintf(out, "Deleted image: %s\n", *i.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
