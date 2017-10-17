///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"io"
	"time"

	"golang.org/x/net/context"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	baseimage "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client/base_image"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	getBaseImagesLong = i18n.T(`Get base images.`)

	// TODO: add examples
	getBaseImagesExample = i18n.T(``)
)

// NewCmdGetBaseImage creates command responsible for getting base images.
func NewCmdGetBaseImage(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "base-image [IMAGE_NAME]",
		Short:   i18n.T("Get base images"),
		Long:    getBaseImagesLong,
		Example: getBaseImagesExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"base-images"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) > 0 {
				err = getBaseImage(out, errOut, cmd, args)
			} else {
				err = getBaseImages(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getBaseImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := imageManagerClient()
	params := &baseimage.GetBaseImageByNameParams{
		Context:       context.Background(),
		BaseImageName: args[0],
	}
	resp, err := client.BaseImage.GetBaseImageByName(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatBaseImageOutput(out, false, []*models.BaseImage{resp.Payload})
}

func getBaseImages(out, errOut io.Writer, cmd *cobra.Command) error {
	client := imageManagerClient()
	params := &baseimage.GetBaseImagesParams{
		Context: context.Background(),
	}
	resp, err := client.BaseImage.GetBaseImages(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatBaseImageOutput(out, true, resp.Payload)
}

func formatBaseImageOutput(out io.Writer, list bool, images []*models.BaseImage) error {
	if vsConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(images)
		}
		return encoder.Encode(images[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "URL", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, image := range images {
		table.Append([]string{*image.Name, *image.DockerURL, string(image.Status), time.Unix(image.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
