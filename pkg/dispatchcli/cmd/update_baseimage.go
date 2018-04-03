///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/base_image"
	"github.com/vmware/dispatch/pkg/image-manager/gen/models"
)

var (
	updateBaseImageLong    = "Updates a base image from a given json representation"
	updateBaseImageExample = `{
		"dockerUrl": "vmware/dispatch-nodejs6-base:0.0.1-dev1
		"groups": null,
		"language": "nodejs6",
		"name": "nodejs6-base",
		"public": true,
		"reason": null,
		"tags": [
			{
				"key": "role",
				"value": "test"
			}
		]
	}`
	imageURL string
)

// CallUpdateBaseImage updates a base image
func CallUpdateBaseImage(input interface{}) error {
	baseImage := input.(*models.BaseImage)
	params := base_image.NewUpdateBaseImageByNameParams()
	params.BaseImageName = *baseImage.Name
	params.Body = baseImage
	_, err := imageManagerClient().BaseImage.UpdateBaseImageByName(params, GetAuthInfoWriter())

	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// NewCmdUpdateBaseImage creates command for updating the base image
func NewCmdUpdateBaseImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "base-image BASE_IMAGE_NAME [--image-url IMAGE_URL] [--language LANGUAGE]",
		Short:   i18n.T("Update base image"),
		Long:    updateBaseImageLong,
		Example: updateBaseImageExample,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updateBaseImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVar(&imageURL, "image-url", "", "The url for the container image.")
	cmd.Flags().StringVar(&language, "language", "", "Specify the runtime language for the image")
	return cmd
}

func updateBaseImage(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {
	baseImageName := args[0]

	params := base_image.NewGetBaseImageByNameParams()
	params.BaseImageName = baseImageName

	ok, err := imageManagerClient().BaseImage.GetBaseImageByName(params, GetAuthInfoWriter())
	if err != nil {
		return errors.Wrapf(err, "Failed retrieving base-image named : %s", baseImageName)
	}

	baseImage := *ok.Payload
	baseImage.Name = &baseImageName
	if cmd.Flags().Changed("image-url") {
		baseImage.DockerURL = &imageURL
	}

	if cmd.Flags().Changed("language") {
		baseImage.Language = swag.String(language)
	}

	err = CallUpdateBaseImage(&baseImage)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated base image: %s\n", baseImageName)
	return nil
}
