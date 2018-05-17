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
	"net"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/toqueteos/webbrowser"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	loginLong = i18n.T(`Login to VMware Dispatch.`)

	// TODO: Add examples
	loginExample = i18n.T(``)
	loginDebug   = false
)

// NewCmdLogin creates a command to login to VMware Dispatch.
func NewCmdLogin(in io.Reader, out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "login",
		Short:   i18n.T("Login to VMware Dispatch."),
		Long:    loginLong,
		Example: loginExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := login(in, out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().BoolVar(&loginDebug, "debug", false, "Extra debug output")

	return cmd
}

const (
	localServerPath  = "/catcher"
	remoteServerPath = "/v1/iam/redirect"
	oauth2Path       = "/v1/iam/oauth2/start"
)

var cookieChan = make(chan string, 1)

func startLocalServer() string {
	server := &http.Server{}
	http.HandleFunc(localServerPath, func(w http.ResponseWriter, req *http.Request) {
		values := req.URL.Query()
		cookie := values.Get("cookie")
		if cookie == "" {
			io.WriteString(w, "Invalid/Error Authorization Cookie.\n")
			cookieChan <- ""
		} else {
			io.WriteString(w, "Cookie received. Please close this page.\n")
			cookieChan <- cookie
		}
	})

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fmt.Printf("LocalServer: Listen() error: %s\n", err)
	}
	go func() {
		err = server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("LocalServer: ListenAndServe() error: %s\n", err)
		}
	}()
	return listener.Addr().String()
}

func login(in io.Reader, out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	if jwtToken != "" ||
		(serviceAccount != "" && jwtPrivateKey != "") {
		return serviceAccountLogin(in, out, errOut, cmd, args)
	}

	return oidcLogin(in, out, errOut, cmd, args)
}

// login Dispatch by OIDC
func oidcLogin(in io.Reader, out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	localServerHost := startLocalServer()
	localServerURI := fmt.Sprintf("http://%s%s", localServerHost, localServerPath)

	// note: two redirects involve here.
	// first, user get authenticated at OAuth2 endpoint e.g. /oauth2/start
	// and if authenticated, will be redirect to identity manager (iam) redirect endpoint
	// second, the redirect endpoint retrieves the cookie
	// redirect the request to the local server, with cookie as a http parameter
	vals := url.Values{
		"rd": {
			fmt.Sprintf("%s?%s", remoteServerPath, url.Values{
				"redirect": {localServerURI},
			}.Encode()),
		},
	}
	requestURL := fmt.Sprintf("https://%s%s?%s", dispatchConfig.Host, oauth2Path, vals.Encode())
	if dispatchConfig.Port != 443 {
		requestURL = fmt.Sprintf("https://%s:%d%s?%s", dispatchConfig.Host, dispatchConfig.Port, oauth2Path, vals.Encode())
	}
	if loginDebug {
		fmt.Fprintf(out, "Logging into: %s\n", requestURL)
	}
	err := webbrowser.Open(requestURL)
	if err != nil {
		return errors.Wrap(err, "error opening web browser")
	}

	cookie := <-cookieChan
	if cookie == "" {
		fmt.Printf("Failed to login, please try again.")
		return nil
	}

	dispatchConfig.Cookie = cookie
	cmdConfig.Contexts[cmdConfig.Current] = &dispatchConfig
	vsConfigJSON, err := json.MarshalIndent(cmdConfig, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error marshalling json")
	}

	err = ioutil.WriteFile(viper.ConfigFileUsed(), vsConfigJSON, 0644)
	if err != nil {
		return errors.Wrapf(err, "error writing configuration to file: %s", viper.ConfigFileUsed())
	}
	fmt.Printf("You have successfully logged in, cookie saved to %s\n", viper.ConfigFileUsed())
	return nil
}

// Login Dispatch by service account
func serviceAccountLogin(in io.Reader, out, errOut io.Writer, cmd *cobra.Command, args []string) (err error) {
	if viper.GetString("dispatchToken") != "" {
		// jwtToken provided
		dispatchConfig.Token = jwtToken
	} else {
		// service acct and private key provided
		dispatchConfig.ServiceAccount = serviceAccount
		dispatchConfig.JWTPrivateKey = jwtPrivateKey
	}

	// write dispatchConfig to file
	vsConfigJSON, err := json.MarshalIndent(dispatchConfig, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error marshalling json")
	}

	err = ioutil.WriteFile(viper.ConfigFileUsed(), vsConfigJSON, 0644)
	if err != nil {
		return errors.Wrapf(err, "error writing configuration to file: %s", viper.ConfigFileUsed())
	}
	fmt.Println("You have successfully logged in!")
	return nil
}
