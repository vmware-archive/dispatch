///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
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
		Run: func(cmd *cobra.Command, args []string) {
			err := login(in, out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

const (
	LocalServerPath  = "/catcher"
	RemoteServerPath = "/v1/iam/redirect"
)

var cookieChan = make(chan string, 1)

func startLocalServer() string {
	server := &http.Server{}
	http.HandleFunc(LocalServerPath, func(w http.ResponseWriter, req *http.Request) {
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

	localServerHost := startLocalServer()
	localServerURI := fmt.Sprintf("http://%s%s", localServerHost, LocalServerPath)
	vals := url.Values{
		"redirect": {localServerURI},
	}

	requestUrl := fmt.Sprintf("https://%s%s?%s", vsConfig.Host, RemoteServerPath, vals.Encode())

	err := webbrowser.Open(requestUrl)
	if err != nil {
		return errors.Wrap(err, "error opening web browser")
	}

	cookie := <-cookieChan
	if cookie == "" {
		fmt.Printf("Failed to login, please try again.")
		return nil
	}

	vsConfig.Cookie = cookie
	vsConfigJson, err := json.MarshalIndent(vsConfig, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error marshalling json")
	}

	err = ioutil.WriteFile(viper.ConfigFileUsed(), vsConfigJson, 0644)
	if err != nil {
		return errors.Wrapf(err, "error writing configuration to file: %s", viper.ConfigFileUsed())
	}
	fmt.Printf("You have successfully logged in, cookie saved to %s\n", viper.ConfigFileUsed())
	return nil
}
