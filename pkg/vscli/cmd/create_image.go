///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"fmt"
	"io"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	apiclient "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client"
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
	host := fmt.Sprintf("%s:%d", vsConfig.Host, vsConfig.Port)
	transport := httptransport.New(host, "/v1/image", []string{"http"})

	client := apiclient.New(transport, strfmt.Default)
	body := &models.Image{
		Name:          &args[0],
		BaseImageName: &args[1],
	}
	params := &image.AddImageParams{
		Body:    body,
		Context: context.Background(),
	}
	created, err := client.Image.AddImage(params)
	if err != nil {
		fmt.Println("create image returned an error")
		clientError, ok := err.(*image.AddImageDefault)
		if ok {
			return fmt.Errorf(*clientError.Payload.Message)
		}
		return err
	}
	fmt.Printf("created image: %s\n", *created.Payload.Name)
	return nil
}
