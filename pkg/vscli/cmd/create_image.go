///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/image"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	createImageLong = i18n.T(`Create serverless image.`)

	// TODO: add examples
	createImageExample = i18n.T(``)
)

// NewCmdCreateImage creates command responsible for image creation.
func NewCmdCreateImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "image IMAGE_NAME BASE_IMAGE_NAME",
		Short:   i18n.T("Create image"),
		Long:    createImageLong,
		Example: createImageExample,
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := createImage(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func createImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := imageManagerClient()
	body := &models.Image{
		Name:          &args[0],
		BaseImageName: &args[1],
	}
	params := &image.AddImageParams{
		Body:    body,
		Context: context.Background(),
	}
	created, err := client.Image.AddImage(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if vsConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*created.Payload)
	}
	fmt.Printf("created image: %s\n", *created.Payload.Name)
	return nil
}
