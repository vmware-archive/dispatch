package apimanager

type VirtualService struct {
	Name     string      `json:"-"`
	Function string      `json:"-"`
	Hosts    []string    `json:"hosts"`
	Gateways []string    `json:"gateways"`
	HTTP     []HttpRoute `json:"http"`
}

type HttpRoute struct {
	Match   []HttpMatch  `json:"match"`
	Route   []DestWeight `json:"route"`
	Rewrite HttpRewrite  `json:"rewrite"`
	CORS    CorsPolicy   `json:"corsPolicy,omitempty"`
}

type HttpMatch struct {
	URI    URI         `json:"uri"`
	Method MethodMatch `json:"method,omitempty"`
}

type URI struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regex  string `json:"regex,omitempty"`
}

type MethodMatch struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regex  string `json:"regex,omitempty"`
}

type DestWeight struct {
	Destination RouteDest `json:"destination"`
	Weight      int       `json:"weight"`
}

type RouteDest struct {
	Host string `json:"host"`
}

type HttpRewrite struct {
	Authority string `json:"authority"`
}

type CorsPolicy struct {
	AllowOrigin  []string `json:"allowOrigin,omitempty"`
	AllowMethods []string `json:"allowMethods,omitempty"`
	AllowHeaders []string `json:"allowHeaders,omitempty"`
}
