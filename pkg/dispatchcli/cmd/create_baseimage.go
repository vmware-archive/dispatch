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
	createBaseImageLong = i18n.T(`Create base image.`)

	// TODO: add examples
	createBaseImageExample = i18n.T(``)
	public                 = false
	language               = i18n.T(``)
)

// NewCmdCreateBaseImage creates command responsible for base image creation.
func NewCmdCreateBaseImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "base-image IMAGE_NAME IMAGE_URL [--public] [--language LANGUAGE]",
		Short:   i18n.T("Create base image"),
		Long:    createBaseImageLong,
		Example: createBaseImageExample,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := createBaseImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVar(&language, "language", "", "Specify the runtime language for the image")
	cmd.Flags().BoolVar(&public, "public", false, "Specify whether the image URL is a public image repository")
	return cmd
}

// CallCreateBaseImage makes the API call to create a base image
func CallCreateBaseImage(bi interface{}) error {
	client := imageManagerClient()
	body := bi.(*models.BaseImage)
	params := &baseimage.AddBaseImageParams{
		Body:    body,
		Context: context.Background(),
	}

	created, err := client.BaseImage.AddBaseImage(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*body = *created.Payload
	return nil
}

func createBaseImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	baseImage := &models.BaseImage{
		Name:      &args[0],
		DockerURL: &args[1],
		Language:  models.Language(language),
		Public:    &public,
	}
	err := CallCreateBaseImage(baseImage)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*baseImage)
	}
	fmt.Fprintf(out, "Created base image: %s\n", *baseImage.Name)
	return nil
}
