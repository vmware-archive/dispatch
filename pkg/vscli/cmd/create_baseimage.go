///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	baseimage "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/base_image"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
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

func createBaseImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := imageManagerClient()
	baseImage := &models.BaseImage{
		Name:      &args[0],
		DockerURL: &args[1],
		Language:  models.Language(language),
		Public:    &public,
	}
	params := &baseimage.AddBaseImageParams{
		Body:    baseImage,
		Context: context.Background(),
	}
	created, err := client.BaseImage.AddBaseImage(params)
	if err != nil {
		return formatAPIError(err, params)
	}
	if vsConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*created.Payload)
	}
	fmt.Fprintf(out, "Created base image: %s\n", *created.Payload.Name)
	return nil
}
