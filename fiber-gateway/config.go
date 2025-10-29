package main

import (
    "encoding/json"
    "os"
    "regexp"
)

type GatewayConfig struct {
    Routes []RouteConfig `json:"routes"`
}

type RouteConfig struct {
    Name         string            `json:"name"`
    Hosts        []string          `json:"hosts,omitempty"`           // host-based routing
    PathPrefix   string            `json:"path_prefix,omitempty"`     // e.g. /api/warrior
    Regex        string            `json:"regex,omitempty"`           // optional regex match
    MethodsAllow []string          `json:"methods_allow,omitempty"`
    MethodsDeny  []string          `json:"methods_deny,omitempty"`
    Upstream     string            `json:"upstream"`                  // primary upstream base URL
    CanaryUp     string            `json:"canary_upstream,omitempty"` // optional canary
    CanaryPct    string            `json:"canary_percent,omitempty"`  // e.g. "10%" or "10"

    // Transformations
    HeadersSet    map[string]string `json:"headers_set,omitempty"`
    HeadersRemove []string          `json:"headers_remove,omitempty"`
    QueryInject   map[string]string `json:"query_inject,omitempty"`
    RewritePrefix string            `json:"rewrite_prefix,omitempty"`  // strip this prefix when proxying

    // Behavior flags
    WebSocketPassthrough bool `json:"websocket_passthrough,omitempty"`
    GrpcProxy            bool `json:"grpc_proxy,omitempty"`           // enable gRPC h2c proxy
}

type compiledRoute struct {
    cfg        RouteConfig
    regex      *regexp.Regexp
    allowSet   map[string]struct{}
    denySet    map[string]struct{}
}

func loadConfig(path string) (*GatewayConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg GatewayConfig
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}

func compileRoutes(cfg *GatewayConfig) ([]compiledRoute, error) {
    out := make([]compiledRoute, 0, len(cfg.Routes))
    for _, r := range cfg.Routes {
        cr := compiledRoute{cfg: r}
        if r.Regex != "" {
            rx, err := regexp.Compile(r.Regex)
            if err != nil {
                return nil, err
            }
            cr.regex = rx
        }
        if len(r.MethodsAllow) > 0 {
            cr.allowSet = make(map[string]struct{}, len(r.MethodsAllow))
            for _, m := range r.MethodsAllow { cr.allowSet[m] = struct{}{} }
        }
        if len(r.MethodsDeny) > 0 {
            cr.denySet = make(map[string]struct{}, len(r.MethodsDeny))
            for _, m := range r.MethodsDeny { cr.denySet[m] = struct{}{} }
        }
        out = append(out, cr)
    }
    return out, nil
}


