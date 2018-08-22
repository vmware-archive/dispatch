package apimanager

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	"istio.io/istio/pilot/pkg/model"

	"github.com/ghodss/yaml"
	"github.com/vmware/dispatch/pkg/api-manager/gen/restapi/operations/endpoint"
	"github.com/vmware/dispatch/pkg/api-manager/istio"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/trace"
)

const (
	// InternalGateway represents the gateway that we should redirect to
	// Should probably come from a config file
	InternalGateway = "knative-ingressgateway.istio-system.svc.cluster.local"
)

func makeMethodRegex(methods []string) string {
	return strings.Join(methods, "|") + "/g"
}

func makeMethodsFromRegex(regex string) []string {
	split := strings.Split(regex, "|")
	return split[0 : len(split)-1]
}

func apiModelOntoIstioEntity(m *v1.API) VirtualService {
	cors := CorsPolicy{}
	if m.Cors {
		cors.AllowMethods = m.Methods
	}
	var routes []HttpRoute
	for _, prefix := range m.Uris {
		if prefix == "/" {
			prefix = fmt.Sprintf("/%v", *m.Function)
		}
		match := HttpMatch{
			URI:    URI{Prefix: prefix},
			Method: MethodMatch{Regex: makeMethodRegex(m.Methods)},
		}
		route := HttpRoute{
			Match: []HttpMatch{match},
			Route: []DestWeight{DestWeight{
				Destination: RouteDest{
					Host: InternalGateway,
				},
				Weight: 100,
			}},
			Rewrite: HttpRewrite{
				Authority: fmt.Sprintf("%v.default.example.com", *m.Function),
			},
			CORS: cors,
		}
		routes = append(routes, route)
	}

	hosts := m.Hosts
	if len(m.Hosts) == 0 {
		hosts = append(hosts, "*")
	}

	return VirtualService{
		Name:     *m.Name,
		Function: *m.Function,
		Hosts:    hosts,
		Gateways: []string{"knative-shared-gateway.knative-serving.svc.cluster.local"},
		HTTP:     routes,
	}
}

func corsEnabled(cors CorsPolicy) bool {
	if len(cors.AllowOrigin) != 0 {
		return true
	} else if len(cors.AllowMethods) != 0 {
		return true
	} else if len(cors.AllowHeaders) != 0 {
		return true
	}
	return false
}

func istioEntityToModel(vs VirtualService) *v1.API {
	var uris []string
	var methods []string
	var function string
	var cors bool
	hosts := vs.Hosts
	for _, route := range vs.HTTP {
		log.Infof("Route: %+v", route)
		uris = append(uris, route.Match[0].URI.Prefix)
		methods = makeMethodsFromRegex(route.Match[0].Method.Regex)
		function = strings.Split(route.Rewrite.Authority, ".")[0]
		if corsEnabled(route.CORS) {
			cors = true
		}
	}
	log.Infof("Virtual Service: %+v", vs)
	log.Infof("Methods: %v", methods)
	return &v1.API{
		Enabled:   true,
		Methods:   methods,
		Name:      &vs.Name,
		Uris:      uris,
		Hosts:     hosts,
		Function:  &function,
		Protocols: []string{"http"},
		Cors:      cors,
	}
}

type IstioHandlers struct {
	client *istio.Client
}

func NewIstioHandlers(client *istio.Client) *IstioHandlers {
	return &IstioHandlers{
		client: client,
	}
}

func (cl *IstioHandlers) AddAPI(params endpoint.AddAPIParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	log.Infof("Trying to add istio api: %+v", params)

	api := apiModelOntoIstioEntity(params.Body)
	virtualServiceSpec, err := yaml.Marshal(api)
	if err != nil {
		log.Errorf("Failed to marshal virtualservicspec: %+v", api)
	}
	err = cl.client.AddAPI(ctx, string(virtualServiceSpec), api.Name)
	if err != nil {
		log.Errorf("Istio failed to create the api: %v", err)
	}
	m := istioEntityToModel(api)
	log.Infof("Added api: %+v", m)
	return endpoint.NewAddAPIOK().WithPayload(m)
}

func (cl *IstioHandlers) GetAPI(params endpoint.GetAPIParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	log.Infof("Trying to get an istio api: %+v", params)
	name := params.API
	virtualService, err := cl.client.GetAPI(ctx, name)
	if err != nil {
		log.Errorf("Couldn't get api %v: %v", name, err)
		return endpoint.NewGetApisDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("API Not found"),
			})
	}

	str, err := model.ToYAML(virtualService.Spec)
	if err != nil {
		return endpoint.NewGetApisDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("Conversion error when getting api"),
			})
	}
	log.Infof("Grabbed istio config: %s", str)

	var api VirtualService
	err = yaml.Unmarshal([]byte(str), &api)
	if err != nil {
		log.Errorf("Unable to unmarshal get result into vs: %v", err)
		return endpoint.NewGetApisDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("Conversion error when getting api"),
			})
	}

	api.Name = virtualService.Name

	m := istioEntityToModel(api)
	m.Enabled = true
	return endpoint.NewGetAPIOK().WithPayload(m)
}

func (cl *IstioHandlers) GetAPIs(params endpoint.GetApisParams, principal interface{}) middleware.Responder {
	log.Errorf("Not ready to get multiple apis")
	return nil
}

func (cl *IstioHandlers) UpdateAPI(params endpoint.UpdateAPIParams, principal interface{}) middleware.Responder {
	log.Infof("Trying to update: %+v", params)
	return cl.AddAPI(endpoint.AddAPIParams{
		HTTPRequest:  params.HTTPRequest,
		XDispatchOrg: params.XDispatchOrg,
		Body:         params.Body,
	}, principal)
}

func (cl *IstioHandlers) DeleteAPI(params endpoint.DeleteAPIParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "")
	defer span.Finish()

	log.Infof("Trying to get an istio api: %+v", params)
	name := params.API
	err := cl.client.DeleteAPI(ctx, name)
	if err != nil {
		log.Errorf("Couldn't delete api %s: %v", name, err)
		return endpoint.NewDeleteAPIDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("Failed to delete api"),
			})
	}

	log.Infof("Successfully deleted api %v", name)
	api := &v1.API{
		Name: &name,
	}

	return endpoint.NewDeleteAPIOK().WithPayload(api)
}
