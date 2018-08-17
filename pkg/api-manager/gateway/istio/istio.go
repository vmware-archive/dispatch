package istio

import (
	"context"
	"fmt"
	"strings"

	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	// InternalGateway represents the gateway that we should redirect to
	// Should probably come from a config file
	InternalGateway = "knative-ingressgateway.istio-system.svc.cluster.local"
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

func fromAPItoConfig(gateway, service string, api *gateway.API) (string, error) {

	if len(api.Hosts) == 0 {
		api.Hosts = []string{"*"}
	}

	var matches []httpMatch
	for _, link := range api.URIs {
		matches = append(matches, httpMatch{
			URI: uri{
				Prefix: link,
			},
		})
	}

	destinationRoute := route{
		Destination: routeDest{
			Host: InternalGateway,
		},
		Weight: 100,
	}

	matcher := httpMatcher{
		Match: matches,
		Rewrite: rewrite{
			Authority: fmt.Sprintf("%v.default.example.com", service),
		},
		Route: []route{destinationRoute},
	}

	if api.CORS {
		matcher.CORS = corsPolicy{
			AllowHeaders: []string{"*"},
			AllowMethods: api.Methods,
		}
	}
	newRoute := reRouterSpec{
		Gateways: []string{gateway},
		Hosts:    api.Hosts,
		HTTP:     []httpMatcher{matcher},
	}
	bytes, err := yaml.Marshal(newRoute)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (c *Client) AddAPI(ctx context.Context, api *gateway.API) (*gateway.API, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	vsSpecs, err := fromAPItoConfig(
		"knative-shared-gateway.knative-serving.svc.cluster.local",
		fmt.Sprintf("%v.default.example.com", api.Function),
		api,
	)
	if err != nil {
		return nil, ewrapper.Wrapf(err, "Unable to build the specs for virtual service")
	}

	vs := model.VirtualService
	spec, err := vs.FromYAML(vsSpecs)
	if err != nil {
		return nil, ewrapper.Wrapf(err, "Unable to build virtual service from spec string")
	}

	cfg := model.Config{
		model.ConfigMeta{
			Type:    "virtual-service",
			Group:   "networking.istio.io",
			Version: "v1alpha3",
			Name:    api.Name,
		},
		spec,
	}

	result, err := c.istioClient.Create(cfg)
	if err != nil {
		return nil, ewrapper.Wrapf(err, "Unable to create virtual service")
	}
	log.Debugf("Created object with resourceVersion: %v", result)

	return api, nil
}

func fromConfigToAPI(cfg *model.Config) *gateway.API {
	spec, _ := model.ToJSONMap(cfg.Spec)
	for field, value := range spec {
		log.Infof("Field: %v, Value: %v\n", field, value)
	}
	return &gateway.API{
		Name: cfg.Name,
	}
}

func (c *Client) GetAPI(ctx context.Context, name string) (*gateway.API, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	cfg, found := c.istioClient.Get("virtual-service", name, "default")
	if !found {
		return nil, ewrapper.Errorf("Istio couldn't located the api %s you requested", name)
	}
	return fromConfigToAPI(cfg), nil
}

func (c *Client) UpdateAPI(ctx context.Context, name string, api *gateway.API) (*gateway.API, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return c.AddAPI(ctx, api)
}

func (c *Client) DeleteAPI(ctx context.Context, api *gateway.API) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	_, _ = c.GetAPI(ctx, api.Name)

	err := c.istioClient.Delete("virtual-service", api.Name, "default")
	if err != nil && !strings.HasSuffix(err.Error(), "not found") {
		return ewrapper.Wrapf(err, "Unable to delete api %+v", api)
	}
	return nil
}
