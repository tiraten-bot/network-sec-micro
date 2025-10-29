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
    Upstream     string            `json:"upstream,omitempty"`        // primary upstream base URL
    Upstreams    []string          `json:"upstreams,omitempty"`       // multiple upstreams for load balancing
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

    // Rate limit
    RateLimit *RateLimitConfig `json:"rate_limit,omitempty"`
    // Circuit breaker
    CircuitBreaker *CircuitBreakerConfig `json:"circuit_breaker,omitempty"`
    // Quota limits
    Quota *QuotaConfig `json:"quota,omitempty"`
    // Header-based routing rules
    HeaderRoutes []HeaderRoute `json:"header_routes,omitempty"`
    // Load balancing policy (round_robin, least_connections)
    LoadBalancing string `json:"load_balancing,omitempty"`
    // Outlier detection
    OutlierDetection *OutlierDetectionConfig `json:"outlier_detection,omitempty"`

    // Aggregation (fan-out/fan-in)
    Aggregates      []AggregateCall      `json:"aggregates,omitempty"`
    AggregatePolicy *AggregatePolicy     `json:"aggregate_policy,omitempty"`

    // Response cache
    Cache *CacheConfig `json:"cache,omitempty"`
}

type RateLimitConfig struct {
    Enabled   bool    `json:"enabled"`
    RPS       float64 `json:"rps"`           // requests per second (for in-memory)
    Burst     int     `json:"burst"`         // burst size (for in-memory)
    KeyHeader string  `json:"key_header"`    // use this header as key; fallback to IP
    UseRedis  bool    `json:"use_redis"`     // enable Redis-backed limiter
    WindowSec int     `json:"window_sec"`    // fixed window (if Redis), seconds
    Limit     int     `json:"limit"`         // max requests per window (if Redis)
}

type CircuitBreakerConfig struct {
    Enabled       bool   `json:"enabled"`
    FailureRatio  float64 `json:"failure_ratio"` // open when failure ratio over window exceeds
    MinRequests   uint32 `json:"min_requests"`
    IntervalSec   int    `json:"interval_sec"`   // moving window
    TimeoutSec    int    `json:"timeout_sec"`    // open state duration
}

type QuotaConfig struct {
    Enabled    bool   `json:"enabled"`
    Daily      int    `json:"daily,omitempty"`
    Hourly     int    `json:"hourly,omitempty"`
    KeyHeader  string `json:"key_header,omitempty"` // keying; fallback IP
}

type HeaderRoute struct {
    Header   string `json:"header"`
    Value    string `json:"value"`
    Upstream string `json:"upstream"`
}

type OutlierDetectionConfig struct {
    Enabled           bool `json:"enabled"`
    FailureThreshold  int  `json:"failure_threshold"`
    EjectDurationSec  int  `json:"eject_duration_sec"`
}

type AggregateCall struct {
    Name     string            `json:"name"`        // key in merged response
    URL      string            `json:"url"`         // absolute or base-relative
    Method   string            `json:"method"`      // GET by default
    TimeoutS int               `json:"timeout_s"`   // per call timeout
    Headers  map[string]string `json:"headers,omitempty"`
}

type AggregatePolicy struct {
    TimeoutS       int    `json:"timeout_s"`         // overall timeout
    Partial        string `json:"partial"`           // ignore|fail_fast
    ResponseRoot   string `json:"response_root"`     // optional root key
    // Simple mapping rules (aggregate-only)
    Rename         map[string]string `json:"rename,omitempty"` // key rename in merged root
    Omit           []string          `json:"omit,omitempty"`
}

type CacheConfig struct {
    Enabled    bool     `json:"enabled"`
    TtlSec     int      `json:"ttl_sec"`
    VaryHeaders []string `json:"vary_headers,omitempty"`
}

type compiledRoute struct {
    cfg        RouteConfig
    regex      *regexp.Regexp
    allowSet   map[string]struct{}
    denySet    map[string]struct{}
    lbIndex    int
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


