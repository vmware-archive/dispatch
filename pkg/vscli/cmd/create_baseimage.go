///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"fmt"
	"io"

	"golang.org/x/net/context"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"

	apiclient "gitlab.eng.vmware.com/serverless/serverless/pkg/image-manager/gen/client"
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
		Args:    cobra.MinimumNArgs(2),
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
	host := fmt.Sprintf("%s:%d", vsConfig.Host, vsConfig.Port)
	transport := httptransport.New(host, "/v1/image", []string{"http"})

	client := apiclient.New(transport, strfmt.Default)
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
		fmt.Println("create base image returned an error")
		return err
	}
	fmt.Printf("created base image: %s\n", *created.Payload.Name)
	return nil
}
