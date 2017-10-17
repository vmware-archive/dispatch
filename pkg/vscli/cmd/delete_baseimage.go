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

func deleteBaseImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := imageManagerClient()
	params := &baseimage.DeleteBaseImageByNameParams{
		Context:       context.Background(),
		BaseImageName: args[0],
	}
	resp, err := client.BaseImage.DeleteBaseImageByName(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatDeleteBaseImageOutput(out, false, []*models.BaseImage{resp.Payload})
}

func formatDeleteBaseImageOutput(out io.Writer, list bool, images []*models.BaseImage) error {
	if vsConfig.Json {
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
