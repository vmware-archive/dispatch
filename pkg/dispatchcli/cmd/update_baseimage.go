///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/base_image"
	"github.com/vmware/dispatch/pkg/image-manager/gen/models"
)

var (
	updateBaseImageLong    = "Updates a base image from a given json representation"
	updateBaseImageExample = `{
		"dockerUrl": "vmware/dispatch-openfaas-nodejs6-base:0.0.3-dev1",
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
	updateFilePath          string
	imageURLUpdateValue     []string
	publicFlagUpdateValue   []string
	languageFlagUpdateValue []string
)

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

func NewCmdUpdateBaseImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "base-image BASE_IMAGE_NAME [--base-image-file BASE_IMAGE_FILE] [--image-url IMAGE_URL] [--public] [--language LANGUAGE]",
		Short:   i18n.T("Update base image"),
		Long:    updateBaseImageLong,
		Example: updateBaseImageExample,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updateBaseImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVar(&updateFilePath, "base-image-file", "", "A json file to update the base image")
	cmd.Flags().StringArrayVar(&imageURLUpdateValue, "image-url", []string{}, "The url for the container image.")
	cmd.Flags().StringArrayVar(&publicFlagUpdateValue, "public", []string{}, "Whether the base image is public or private")
	cmd.Flags().StringArrayVar(&languageFlagUpdateValue, "language", []string{}, "Specify the runtime language for the image")
	return cmd
}

func updateBaseImage(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {
	baseImageName := args[0]
	var baseImage = models.BaseImage{}

	if updateFilePath != "" && (len(imageURLUpdateValue) != 0 || len(publicFlagUpdateValue) != 0 || len(languageFlagUpdateValue) != 0) {
		message := fmt.Sprintf("base-image-file flag cannot be used with any other flag")
		return formatCliError(errors.New("Input error"), message)
	}

	if updateFilePath != "" {
		baseImageContent, err := ioutil.ReadFile(updateFilePath)
		if err != nil {
			message := fmt.Sprintf("Error when reading content of %s", updateFilePath)
			return formatCliError(err, message)
		}
		if err := json.Unmarshal(baseImageContent, &baseImage); err != nil {
			message := fmt.Sprintf("Error when parsing JSON from %s with error %s", baseImageContent, err)
			return formatCliError(err, message)
		}
	}

	if baseImageName != *baseImage.Name {
		message := fmt.Sprintf("BASE_IMAGE_NAME does not match name in base image json.")
		return formatCliError(errors.New("Input mismatch"), message)
	}

	err := CallUpdateBaseImage(&baseImage)

	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated base image: %s\n", *baseImage.Name)
	return nil
}
