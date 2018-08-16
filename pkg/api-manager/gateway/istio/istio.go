package istio

import (
	"context"
	"flag"
	"fmt"
	"log"

	"istio.io/istio/pilot/pkg/config/kube/crd"
	// "istio.io/istio/pilot/pkg/model"

	ewrapper "github.com/pkg/errors"

	"github.com/vmware/dispatch/pkg/api-manager/gateway"
)

func NewClient() (*Client, error) {
	crdClient, err := crd.NewClient("", "", []model.ProtoSchema{model.VirtualService}, "")
	if err != nil {
		return nil, err
	}
	return &Client{
		istioClient: crdClient,
	}
}

type Client struct {
	istioClient *crd.Client
}

func (c *Client) AddAPI(ctx context.Context, api *API) (*API, error) {
	return nil, ewrapper.Errorf("Not implemented yet dummy")
}
func (c *Client) GetAPI(ctx context.Context, name string) (*API, error) {
	return nil, ewrapper.Errorf("Not implemented yet dummy")
}
func (c *Client) UpdateAPI(ctx context.Context, name string, api *API) (*API, error) {
	return nil, ewrapper.Errorf("Not implemented yet dummy")
}
func (c *Client) DeleteAPI(ctx context.Context, api *API) error {
	return nil, ewrapper.Errorf("Not implemented yet dummy")
}
