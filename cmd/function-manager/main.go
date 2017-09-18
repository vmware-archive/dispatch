///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////

package main

import (
	"log"
	"os"

	loads "github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	flags "github.com/jessevdk/go-flags"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions/gen/restapi"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/functions/gen/restapi/operations"
)

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

func main() {

	server := restapi.NewServer(nil)

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Function manager"
	parser.LongDescription = "This is the API server for the serverless function manager service.\n"

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	swaggerSpec, err := loads.Spec(string(server.Spec))
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewFunctionManagerAPI(swaggerSpec)
	server.SetAPI(api)
	defer server.Shutdown()

	server.ConfigureAPI()

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}
