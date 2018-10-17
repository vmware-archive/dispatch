///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	pkgUtils "github.com/vmware/dispatch/pkg/utils"
)

var (
	editLong = i18n.T(`Edit a resource. The resource type should be one of 
[api|base-image|image|function|eventdriver|eventdrivertype|subscription|secret|service|organization|policy]
For update function, you must specify the new sorucepath`)

	// TODO: Add examples
	editExample = i18n.T(``)
)

// NewCmdEdit edit command responsible for resource edit.
func NewCmdEdit(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit TYPE [NAME|ID]",
		Short:   i18n.T("Edit resources."),
		Long:    editLong,
		Example: editExample,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				runHelp(cmd, args)
				return
			}
			imgClient := imageManagerClient()
			eventClient := eventManagerClient()
			apiClient := apiManagerClient()
			secClient := secretStoreClient()
			iamClient := identityManagerClient()
			fnClient := functionManagerClient()

			updateMap := map[string]ModelAction{
				pkgUtils.APIKind:            CallUpdateAPI(apiClient),
				pkgUtils.ApplicationKind:    CallUpdateApplication,
				pkgUtils.BaseImageKind:      CallUpdateBaseImage(imgClient),
				pkgUtils.DriverKind:         CallUpdateDriver(eventClient),
				pkgUtils.DriverTypeKind:     CallUpdateDriverType(eventClient),
				pkgUtils.FunctionKind:       CallUpdateFunction(fnClient),
				pkgUtils.ImageKind:          CallUpdateImage(imgClient),
				pkgUtils.SecretKind:         CallUpdateSecret(secClient),
				pkgUtils.SubscriptionKind:   CallUpdateSubscription(eventClient),
				pkgUtils.PolicyKind:         CallUpdatePolicy(iamClient),
				pkgUtils.ServiceAccountKind: CallUpdateServiceAccount(iamClient),
				pkgUtils.OrganizationKind:   CallUpdateOrganization(iamClient),
			}

			buf := new(bytes.Buffer)
			encoder := yaml.NewEncoder(buf)
			var resp interface{}
			var err error
			switch args[0] {
			case "function":
				resp, err = fnClient.GetFunction(context.TODO(), dispatchConfig.Organization, args[1])
			case "base-image":
				resp, err = imgClient.GetBaseImage(context.TODO(), dispatchConfig.Organization, args[1])
			case "image":
				resp, err = imgClient.GetImage(context.TODO(), dispatchConfig.Organization, args[1])
			case "api":
				resp, err = apiClient.GetAPI(context.TODO(), dispatchConfig.Organization, args[1])
			case "eventdriver":
				resp, err = eventClient.GetEventDriver(context.TODO(), dispatchConfig.Organization, args[1])
			case "eventdrivertype":
				resp, err = eventClient.GetEventDriverType(context.TODO(), dispatchConfig.Organization, args[1])
			case "secret":
				resp, err = secClient.GetSecret(context.TODO(), dispatchConfig.Organization, args[1])
			case "subscription":
				resp, err = eventClient.GetSubscription(context.TODO(), dispatchConfig.Organization, args[1])
			case "policy":
				resp, err = iamClient.GetPolicy(context.TODO(), dispatchConfig.Organization, args[1])
			case "organization":
				resp, err = iamClient.GetOrganization(context.TODO(), dispatchConfig.Organization, args[1])
			case "service":
				resp, err = iamClient.GetServiceAccount(context.TODO(), dispatchConfig.Organization, args[1])
			default:
				runHelp(cmd, args)
				return
			}
			CheckErr(err)
			encoder.Encode(resp)
			fpath, modify, err := editFile(buf)
			CheckErr(err)
			if modify {
				err = applyUpdate(out, updateMap, "Updated", fpath)
				CheckErr(err)
			}
			return
		},
	}

	cmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Working directory relative paths are based on")

	return cmd
}

//RandString generate a random string
func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func editFile(buf *bytes.Buffer) (string, bool, error) {
	fpath := os.TempDir() + RandString(16)
	f, err := os.Create(fpath)
	if err != nil {
		return fpath, false, err
	}
	fmt.Fprint(f, buf.String())
	f.Close()
	tmpFileSha256_1, err := getSha256(fpath)
	if err != nil {
		return fpath, false, err
	}
	editorName := os.Getenv("EDITOR")
	if len(editorName) == 0 {
		editorName = "vi"
	}
	editor := exec.Command(editorName, fpath)
	editor.Stdin = os.Stdin
	editor.Stdout = os.Stdout
	editor.Stderr = os.Stderr
	err = editor.Start()
	if err != nil {
		return fpath, false, err
	}
	err = editor.Wait()
	if err != nil {
		return fpath, false, err
	}
	tmpFileSha256_2, err := getSha256(fpath)
	if err != nil {
		return fpath, false, err
	}
	if tmpFileSha256_1 == tmpFileSha256_2 {
		fmt.Println("Edit cancelled, no changes made.")
		return fpath, false, nil
	}
	return fpath, true, nil
}

func applyUpdate(out io.Writer, actionMap map[string]ModelAction, actionName string, file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "Error reading file %s", file)
	}
	return importBytes(out, b, actionMap, actionName)
}

func getSha256(fpath string) (string, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
