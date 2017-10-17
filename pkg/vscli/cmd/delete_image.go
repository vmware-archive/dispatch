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

	image "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/image"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
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

func deleteImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := imageManagerClient()
	params := &image.DeleteImageByNameParams{
		Context:   context.Background(),
		ImageName: args[0],
	}
	resp, err := client.Image.DeleteImageByName(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatDeleteImageOutput(out, false, []*models.Image{resp.Payload})
}

func formatDeleteImageOutput(out io.Writer, list bool, images []*models.Image) error {
	if vsConfig.Json {
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
