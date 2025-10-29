package main

import (
    "log"
    "net"
    "net/http"
    "net/http/httputil"
    "os"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/proxy"
    adaptor "github.com/gofiber/adaptor/v2"
    "golang.org/x/net/http2"
    "golang.org/x/time/rate"
    "github.com/sony/gobreaker"
    "crypto/tls"
    "sync"
)

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func main() {
    // Fiber app with timeouts and body limit (basic limits)
    app := fiber.New(fiber.Config{
        BodyLimit:      10 * 1024 * 1024, // 10MB
        ReadTimeout:    15 * time.Second,
        WriteTimeout:   30 * time.Second,
        IdleTimeout:    60 * time.Second,
        Prefork:        false,
        CaseSensitive:  true,
        StrictRouting:  false,
        EnablePrintRoutes: false,
    })

    // Basic logging
    app.Use(logger.New())

    // CORS (can be restricted later per route)
    app.Use(cors.New(cors.Config{
        AllowOrigins:     getEnv("CORS_ALLOW_ORIGINS", "*"),
        AllowMethods:     getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
        AllowHeaders:     getEnv("CORS_ALLOW_HEADERS", "Authorization,Content-Type,Accept,X-Request-ID"),
        ExposeHeaders:    getEnv("CORS_EXPOSE_HEADERS", "X-Request-ID"),
        AllowCredentials: true,
        MaxAge:           300,
    }))

    // Request transform: inject correlation id if missing
    app.Use(func(c *fiber.Ctx) error {
        if c.Get("X-Request-ID") == "" {
            c.Request().Header.Add("X-Request-ID", time.Now().UTC().Format(time.RFC3339Nano))
        }
        return c.Next()
    })

    // Method policy example: deny TRACE, limit others later per route
    app.Use(func(c *fiber.Ctx) error {
        if c.Method() == fiber.MethodTrace {
            return c.Status(fiber.StatusMethodNotAllowed).SendString("method not allowed")
        }
        return c.Next()
    })

    // Try declarative config first
    if cfgPath := os.Getenv("GW_CONFIG"); cfgPath != "" {
        if cfg, err := loadConfig(cfgPath); err == nil {
            if croutes, err := compileRoutes(cfg); err == nil {
                // Single catch-all; apply first-match policy per config order
                // per-route state
                type routeState struct {
                    rlMu sync.Mutex
                    rlMap map[string]*rate.Limiter
                    cb    *gobreaker.CircuitBreaker
                    lbMu  sync.Mutex
                    lbIdx int
                }
                states := make([]routeState, len(croutes))

                // init circuit breakers
                for i := range croutes {
                    rc := croutes[i].cfg
                    if rc.CircuitBreaker != nil && rc.CircuitBreaker.Enabled {
                        st := &states[i]
                        st.cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
                            Name:        rc.Name,
                            Interval:    time.Duration(rc.CircuitBreaker.IntervalSec) * time.Second,
                            Timeout:     time.Duration(rc.CircuitBreaker.TimeoutSec) * time.Second,
                            ReadyToTrip: func(counts gobreaker.Counts) bool {
                                if counts.Requests < rc.CircuitBreaker.MinRequests {
                                    return false
                                }
                                failRatio := 0.0
                                if counts.Requests > 0 {
                                    failRatio = float64(counts.TotalFailures) / float64(counts.Requests)
                                }
                                return failRatio >= rc.CircuitBreaker.FailureRatio
                            },
                        })
                    }
                    states[i].rlMap = make(map[string]*rate.Limiter)
                }

                app.All("/*", func(c *fiber.Ctx) error {
                    host := c.Hostname()
                    path := c.OriginalURL()
                    method := c.Method()

                    for idx, cr := range croutes {
                        st := &states[idx]
                        // host match (if provided)
                        if len(cr.cfg.Hosts) > 0 {
                            matched := false
                            for _, h := range cr.cfg.Hosts {
                                if h == host { matched = true; break }
                            }
                            if !matched { continue }
                        }
                        // path prefix match (if provided)
                        if cr.cfg.PathPrefix != "" {
                            if !hasPrefix(path, cr.cfg.PathPrefix) { continue }
                        }
                        // regex match (if provided)
                        if cr.regex != nil && !cr.regex.MatchString(path) { continue }
                        // method policy
                        if len(cr.denySet) > 0 {
                            if _, deny := cr.denySet[method]; deny { return c.SendStatus(fiber.StatusMethodNotAllowed) }
                        }
                        if len(cr.allowSet) > 0 {
                            if _, allow := cr.allowSet[method]; !allow { return c.SendStatus(fiber.StatusMethodNotAllowed) }
                        }

                        // rate limit per key
                        if cr.cfg.RateLimit != nil && cr.cfg.RateLimit.Enabled {
                            key := c.IP()
                            if kh := cr.cfg.RateLimit.KeyHeader; kh != "" {
                                if v := c.Get(kh); v != "" { key = v }
                            }
                            st.rlMu.Lock()
                            lim, ok := st.rlMap[key]
                            if !ok {
                                rps := cr.cfg.RateLimit.RPS
                                if rps <= 0 { rps = 10 }
                                burst := cr.cfg.RateLimit.Burst
                                if burst <= 0 { burst = int(rps) }
                                lim = rate.NewLimiter(rate.Limit(rps), burst)
                                st.rlMap[key] = lim
                            }
                            st.rlMu.Unlock()
                            if !lim.Allow() {
                                return c.Status(fiber.StatusTooManyRequests).SendString("rate limit exceeded")
                            }
                        }

                        // apply transforms
                        // headers set
                        for k, v := range cr.cfg.HeadersSet { c.Request().Header.Set(k, v) }
                        // headers remove
                        for _, k := range cr.cfg.HeadersRemove { c.Request().Header.Del(k) }
                        // query inject
                        if len(cr.cfg.QueryInject) > 0 {
                            q := c.Request().URI().QueryArgs()
                            for k, v := range cr.cfg.QueryInject { q.Set(k, v) }
                        }

                        // load balancing pick
                        base := cr.cfg.Upstream
                        if len(cr.cfg.Upstreams) > 0 {
                            st.lbMu.Lock()
                            base = cr.cfg.Upstreams[st.lbIdx%len(cr.cfg.Upstreams)]
                            st.lbIdx++
                            st.lbMu.Unlock()
                        }
                        // canary selection
                        if cr.cfg.CanaryUp != "" && cr.cfg.CanaryPct != "" {
                            rid := c.Get("X-Request-ID")
                            if len(rid) > 0 {
                                last := rid[len(rid)-1]
                                pct := 10
                                if p, err := parsePercent(cr.cfg.CanaryPct); err == nil { pct = p }
                                if int(last)%100 < pct { base = cr.cfg.CanaryUp }
                            }
                        }

                        // rewrite
                        targetPath := path
                        if cr.cfg.RewritePrefix != "" && hasPrefix(path, cr.cfg.RewritePrefix) {
                            targetPath = path[len(cr.cfg.RewritePrefix):]
                            if !hasPrefix(targetPath, "/") { targetPath = "/" + targetPath }
                        }

                        // websocket passthrough: upgrade to ws(s) if needed
                        if cr.cfg.WebSocketPassthrough && isWebSocket(c) {
                            if st.cb != nil {
                                // short-circuit if breaker open
                                if st.cb.State() == gobreaker.StateOpen {
                                    return c.Status(fiber.StatusServiceUnavailable).SendString("circuit open")
                                }
                            }
                            target := toWS(base) + targetPath
                            return proxy.Do(c, target)
                        }

                        // gRPC proxy via h2c reverse proxy (net/http) mounted in Fiber
                        if cr.cfg.GrpcProxy {
                            if st.cb != nil && st.cb.State() == gobreaker.StateOpen {
                                return c.Status(fiber.StatusServiceUnavailable).SendString("circuit open")
                            }
                            h := newH2cReverseProxy(base, cr.cfg.RewritePrefix)
                            if st.cb == nil {
                                return adaptor.HTTPHandler(h)(c)
                            }
                            // run via breaker
                            _, err := st.cb.Execute(func() (interface{}, error) {
                                return nil, adaptor.HTTPHandler(h)(c)
                            })
                            return err
                        }

                        // default http proxy
                        target := base + targetPath
                        if st.cb == nil {
                            return proxy.Do(c, target)
                        }
                        _, err := st.cb.Execute(func() (interface{}, error) {
                            return nil, proxy.Do(c, target)
                        })
                        return err
                    }
                    return c.SendStatus(fiber.StatusNotFound)
                })
            }
        }
    }

    // Fallback static upstreams (if no config provided)
    warriorUp := getEnv("UPSTREAM_WARRIOR", "http://localhost:8080")
    warriorCanaryUp := os.Getenv("UPSTREAM_WARRIOR_CANARY")
    warriorCanaryPct := os.Getenv("UPSTREAM_WARRIOR_CANARY_PERCENT")
    enemyUp := getEnv("UPSTREAM_ENEMY", "http://localhost:8083")
    dragonUp := getEnv("UPSTREAM_DRAGON", "http://localhost:8084")
    weaponUp := getEnv("UPSTREAM_WEAPON", "http://localhost:8081")

    // Health
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok", "service": "fiber-gateway"})
    })

    // L7 routing: path-based proxy groups
    // Warrior service
    app.All("/api/warrior/*", func(c *fiber.Ctx) error {
        targetBase := warriorUp
        if warriorCanaryUp != "" && warriorCanaryPct != "" {
            // coarse canary selection using request-id hash last digit
            rid := c.Get("X-Request-ID")
            if len(rid) > 0 {
                last := rid[len(rid)-1]
                // default 10% if parse fails
                pct := 10
                if p, err := parsePercent(warriorCanaryPct); err == nil {
                    pct = p
                }
                if int(last)%100 < pct {
                    targetBase = warriorCanaryUp
                }
            }
        }
        target := targetBase + c.OriginalURL()[len("/api/warrior"):]
        return proxy.Do(c, target)
    })

    // Enemy service
    app.All("/api/enemy/*", func(c *fiber.Ctx) error {
        target := enemyUp + c.OriginalURL()[len("/api/enemy"):]
        return proxy.Do(c, target)
    })

    // Dragon service
    app.All("/api/dragon/*", func(c *fiber.Ctx) error {
        target := dragonUp + c.OriginalURL()[len("/api/dragon"):]
        return proxy.Do(c, target)
    })

    // Weapon service
    app.All("/api/weapon/*", func(c *fiber.Ctx) error {
        target := weaponUp + c.OriginalURL()[len("/api/weapon"):]
        return proxy.Do(c, target)
    })

    // Regex route example (future use):
    // app.All("/api/(?i)public/.*", someHandler)

    // Start server
    port := getEnv("GW_PORT", "8090")
    log.Printf("fiber-gateway listening on :%s", port)
    if err := app.Listen(":" + port); err != nil {
        log.Fatalf("gateway error: %v", err)
    }
}

func parsePercent(s string) (int, error) {
    // accept values like "10" or "10%"
    if len(s) > 0 && (s[len(s)-1] == '%' || s[len(s)-1] == ' ') {
        s = s[:len(s)-1]
    }
    v := 0
    for i := 0; i < len(s); i++ {
        if s[i] < '0' || s[i] > '9' {
            return 0, fiber.ErrBadRequest
        }
        v = v*10 + int(s[i]-'0')
    }
    if v < 0 {
        v = 0
    }
    if v > 100 {
        v = 100
    }
    return v, nil
}

func hasPrefix(s, p string) bool {
    if len(p) == 0 { return true }
    if len(s) < len(p) { return false }
    return s[:len(p)] == p
}

func isWebSocket(c *fiber.Ctx) bool {
    if c.Get("Upgrade") == "websocket" { return true }
    conn := c.Get("Connection")
    return conn == "Upgrade" || conn == "upgrade"
}

func toWS(httpURL string) string {
    if hasPrefix(httpURL, "https://") { return "wss://" + httpURL[len("https://"):] }
    if hasPrefix(httpURL, "http://") { return "ws://" + httpURL[len("http://"):] }
    return httpURL
}

// newH2cReverseProxy creates a ReverseProxy that supports cleartext HTTP/2 (h2c)
func newH2cReverseProxy(base string, stripPrefix string) http.Handler {
    director := func(r *http.Request) {
        // Build target URL by joining base with request URI minus stripPrefix
        // r.URL already has the original path
        targetPath := r.URL.Path
        if stripPrefix != "" && hasPrefix(targetPath, stripPrefix) {
            targetPath = targetPath[len(stripPrefix):]
            if targetPath == "" || targetPath[0] != '/' {
                targetPath = "/" + targetPath
            }
        }
        // Parse base manually
        // Expect base like http://host:port or https://host:port
        if hasPrefix(base, "http://") {
            r.URL.Scheme = "http"
            r.URL.Host = base[len("http://"):]
        } else if hasPrefix(base, "https://") {
            r.URL.Scheme = "https"
            r.URL.Host = base[len("https://"):]
        } else {
            r.URL.Scheme = "http"
            r.URL.Host = base
        }
        r.URL.Path = targetPath
        // pass through
    }

    rp := &httputil.ReverseProxy{Director: director}

    // h2c transport (allows HTTP/2 without TLS)
    rp.Transport = &http2.Transport{
        AllowHTTP: true,
        DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
            // Use plain TCP for h2c
            return net.Dial(network, addr)
        },
    }
    return rp
}


