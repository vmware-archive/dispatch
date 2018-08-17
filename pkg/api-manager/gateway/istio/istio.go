package istio

import (
	"context"

	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api-manager/gateway"
)

func NewClient() (*Client, error) {
	log.Infoln("Trying to create istio client")
	crdClient, err := crd.NewClient("", "", []model.ProtoSchema{model.VirtualService}, "")
	if err != nil {
		return nil, err
	}
	return &Client{
		istioClient: crdClient,
	}, nil
}

type Client struct {
	istioClient *crd.Client
}

func (c *Client) AddAPI(ctx context.Context, api *gateway.API) (*gateway.API, error) {
	return nil, ewrapper.Errorf("Not implemented yet dummy")
}
func (c *Client) GetAPI(ctx context.Context, name string) (*gateway.API, error) {
	return nil, ewrapper.Errorf("Not implemented yet dummy")
}
func (c *Client) UpdateAPI(ctx context.Context, name string, api *gateway.API) (*gateway.API, error) {
	return nil, ewrapper.Errorf("Not implemented yet dummy")
}
func (c *Client) DeleteAPI(ctx context.Context, api *gateway.API) error {
	return ewrapper.Errorf("Not implemented yet dummy")
}
