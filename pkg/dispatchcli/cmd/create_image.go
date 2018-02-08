///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/image-manager/gen/client/image"
	models "github.com/vmware/dispatch/pkg/image-manager/gen/models"
)

var (
	createImageLong = i18n.T(`Create dispatch image.`)

	// TODO: add examples
	createImageExample      = i18n.T(``)
	systemDependenciesFile  = i18n.T(``)
	runtimeDependenciesFile = i18n.T(``)
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
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringVar(&systemDependenciesFile, "system-deps", "", "path to file with system dependencies")
	cmd.Flags().StringVar(&runtimeDependenciesFile, "runtime-deps", "", "path to file with runtime dependencies")
	return cmd
}

type cliImage struct {
	models.Image
	RuntimeDependenciesPath string `json:"runtimeDependenciesPath"`
	SystemDependenciesPath  string `json:"systemDependenciesPath"`
}

// CallCreateImage makes the API call to create an image
func CallCreateImage(i interface{}) error {
	client := imageManagerClient()
	cli := i.(*cliImage)

	var systemDependencies models.SystemDependencies
	if cli.SystemDependenciesPath != "" {
		fullPath := path.Join(workDir, cli.SystemDependenciesPath)
		b, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("Failed to read system dependencies file: %s", err)
		}
		err = json.Unmarshal(b, &systemDependencies)
		if err != nil {
			return fmt.Errorf("Failed to unmarshal system dependencies file: %s", err)
		}
	}
	var runtimeDependencies models.RuntimeDependencies
	if cli.RuntimeDependenciesPath != "" {
		fullPath := path.Join(workDir, cli.RuntimeDependenciesPath)
		b, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("Failed to read runtime dependencies file: %s", err)
		}
		runtimeDependencies.Manifest = string(b)
	}
	cli.SystemDependencies = &systemDependencies
	cli.RuntimeDependencies = &runtimeDependencies

	params := &image.AddImageParams{
		Body:    &cli.Image,
		Context: context.Background(),
	}

	created, err := client.Image.AddImage(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	cli.Image = *created.Payload
	return nil
}

func createImage(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	cli := &cliImage{
		Image: models.Image{
			Name:          &args[0],
			BaseImageName: &args[1],
			Tags:          models.ImageTags{},
		},
		SystemDependenciesPath:  systemDependenciesFile,
		RuntimeDependenciesPath: runtimeDependenciesFile,
	}
	if cmdFlagApplication != "" {
		cli.Tags = append(cli.Tags, &models.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}
	err := CallCreateImage(cli)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(cli.Image)
	}
	fmt.Fprintf(out, "Created image: %s\n", *cli.Name)
	return nil
}
