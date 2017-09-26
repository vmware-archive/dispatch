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
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"

	apiclient "gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/client"
	authentication "gitlab.eng.vmware.com/serverless/serverless/pkg/identity-manager/gen/client/authentication"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	loginLong = i18n.T(`Login to VMware serverless platform.`)

	// TODO: Add examples
	loginExample = i18n.T(``)
)

// NewCmdLogin creates a command to login to VMware serverless platform.
func NewCmdLogin(in io.Reader, out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "login",
		Short:   i18n.T("Login to VMware serverless platform."),
		Long:    loginLong,
		Example: loginExample,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := login(in, out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func login(in io.Reader, out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	fmt.Fprint(out, "Please Enter Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		fmt.Fprintf(errOut, "password error: %s", err)
	}
	password := string(bytePassword)
	fmt.Fprint(out, "\n")

	username := args[0]
	host := fmt.Sprintf("%s:%d", vsConfig.Host, vsConfig.Port)
	transport := httptransport.New(host, "", []string{"http"})
	client := apiclient.New(transport, strfmt.Default)
	params := &authentication.LoginPasswordParams{
		Username: &username,
		Password: &password,
		Context:  context.Background(),
	}

	body, err := client.Authentication.LoginPassword(params)
	if err != nil {
		// fmt.Println("login returned an error")
		clientError, ok := err.(*authentication.LoginPasswordDefault)
		if ok {
			return fmt.Errorf("log in error: %s", *clientError.Payload.Message)
		}
		return err
	}
	cookie := body.SetCookie
	fmt.Printf("successfully logged in, with cookie: %s\n", cookie)
	return nil
}
