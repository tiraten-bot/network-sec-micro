package main

import (
    "log"
    "sync"
    "time"

    "network-sec-micro/pkg/secrets"

    "github.com/gofiber/adaptor/v2"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    redis "github.com/redis/go-redis/v9"
    "github.com/sony/gobreaker"
    "golang.org/x/time/rate"
)

func getEnv(key, def string) string {
    return secrets.GetOrDefault(key, def)
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
    if cfgPath := secrets.GetOrDefault("GW_CONFIG", ""); cfgPath != "" {
        if cfg, err := loadConfig(cfgPath); err == nil {
            if croutes, err := compileRoutes(cfg); err == nil {
                // Redis client (optional)
                var rdb *redis.Client
                if addr := secrets.GetOrDefault("REDIS_ADDR", ""); addr != "" {
                    rdb = redis.NewClient(&redis.Options{
                        Addr:     addr,
                        Password: secrets.GetOrDefault("REDIS_PASSWORD", ""),
                        DB:       0,
                    })
                }

                // Single catch-all; apply first-match policy per config order
                // per-route state
                type routeState struct {
                    rlMu sync.Mutex
                    rlMap map[string]*rate.Limiter
                    cb    *gobreaker.CircuitBreaker
                    lbMu  sync.Mutex
                    lbIdx int
                    // least-connections & outlier detection
                    active map[string]int
                    fail   map[string]int
                    ejectedUntil map[string]time.Time
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
                    states[i].active = make(map[string]int)
                    states[i].fail = make(map[string]int)
                    states[i].ejectedUntil = make(map[string]time.Time)
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

                        // rate limit per key (Redis fixed-window if configured)
                        if cr.cfg.RateLimit != nil && cr.cfg.RateLimit.Enabled {
                            key := c.IP()
                            if kh := cr.cfg.RateLimit.KeyHeader; kh != "" { if v := c.Get(kh); v != "" { key = v } }
                            if cr.cfg.RateLimit.UseRedis && rdb != nil {
                                win := cr.cfg.RateLimit.WindowSec
                                if win <= 0 { win = 1 }
                                lim := cr.cfg.RateLimit.Limit
                                if lim <= 0 { lim = 100 }
                                now := time.Now().Unix()
                                bucket := now - (now % int64(win))
                                rkey := "rl:" + cr.cfg.Name + ":" + key + ":" + fmtInt(bucket)
                                allowed, err := redisAllow(c.Context(), rdb, rkey, lim, win)
                                if err != nil { return c.Status(fiber.StatusInternalServerError).SendString("rl err") }
                                if !allowed { return c.Status(fiber.StatusTooManyRequests).SendString("rate limit exceeded") }
                            } else {
                                st.rlMu.Lock()
                                lim, ok := st.rlMap[key]
                                if !ok {
                                    rps := cr.cfg.RateLimit.RPS; if rps <= 0 { rps = 10 }
                                    burst := cr.cfg.RateLimit.Burst; if burst <= 0 { burst = int(rps) }
                                    lim = rate.NewLimiter(rate.Limit(rps), burst)
                                    st.rlMap[key] = lim
                                }
                                st.rlMu.Unlock()
                                if !lim.Allow() { return c.Status(fiber.StatusTooManyRequests).SendString("rate limit exceeded") }
                            }
                        }

                        // quota (Redis counters)
                        if cr.cfg.Quota != nil && cr.cfg.Quota.Enabled && rdb != nil {
                            qkey := c.IP(); if kh := cr.cfg.Quota.KeyHeader; kh != "" { if v := c.Get(kh); v != "" { qkey = v } }
                            if cr.cfg.Quota.Hourly > 0 {
                                ts := time.Now().UTC()
                                base := ts.Format("2006010215")
                                rkey := "q:h:" + cr.cfg.Name + ":" + qkey + ":" + base
                                ok, err := redisQuota(c.Context(), rdb, rkey, cr.cfg.Quota.Hourly, int(time.Until(ts.Truncate(time.Hour).Add(time.Hour)).Seconds()))
                                if err != nil { return c.Status(fiber.StatusInternalServerError).SendString("quota err") }
                                if !ok { return c.Status(fiber.StatusTooManyRequests).SendString("hourly quota exceeded") }
                            }
                            if cr.cfg.Quota.Daily > 0 {
                                ts := time.Now().UTC()
                                base := ts.Format("20060102")
                                // seconds until next midnight UTC
                                tomorrow := time.Date(ts.Year(), ts.Month(), ts.Day()+1, 0, 0, 0, 0, time.UTC)
                                ttl := int(time.Until(tomorrow).Seconds())
                                rkey := "q:d:" + cr.cfg.Name + ":" + qkey + ":" + base
                                ok, err := redisQuota(c.Context(), rdb, rkey, cr.cfg.Quota.Daily, ttl)
                                if err != nil { return c.Status(fiber.StatusInternalServerError).SendString("quota err") }
                                if !ok { return c.Status(fiber.StatusTooManyRequests).SendString("daily quota exceeded") }
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

                        // header-based routing override
                        base := cr.cfg.Upstream
                        if len(cr.cfg.HeaderRoutes) > 0 {
                            for _, hr := range cr.cfg.HeaderRoutes {
                                if c.Get(hr.Header) == hr.Value { base = hr.Upstream; break }
                            }
                        }
                        // load balancing pick
                        if len(cr.cfg.Upstreams) > 0 {
                            // skip ejected
                            pick := ""
                            if cr.cfg.LoadBalancing == "least_connections" {
                                st.lbMu.Lock()
                                min := 1<<31 - 1
                                for _, up := range cr.cfg.Upstreams {
                                    if until, ok := st.ejectedUntil[up]; ok && time.Now().Before(until) { continue }
                                    cur := st.active[up]
                                    if cur < min { min = cur; pick = up }
                                }
                                if pick == "" { pick = cr.cfg.Upstreams[st.lbIdx%len(cr.cfg.Upstreams)]; st.lbIdx++ }
                                st.lbMu.Unlock()
                            } else {
                                st.lbMu.Lock()
                                // round robin skipping ejected if possible
                                for i := 0; i < len(cr.cfg.Upstreams); i++ {
                                    candidate := cr.cfg.Upstreams[(st.lbIdx+i)%len(cr.cfg.Upstreams)]
                                    if until, ok := st.ejectedUntil[candidate]; ok && time.Now().Before(until) { continue }
                                    pick = candidate; st.lbIdx = (st.lbIdx+i+1)%len(cr.cfg.Upstreams); break
                                }
                                if pick == "" { pick = cr.cfg.Upstreams[st.lbIdx%len(cr.cfg.Upstreams)]; st.lbIdx++ }
                                st.lbMu.Unlock()
                            }
                            base = pick
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
                            st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                            err := proxy.Do(c, target)
                            st.lbMu.Lock()
                            st.active[base]--
                            if err != nil {
                                st.fail[base]++
                                maybeEject(cr.cfg, st, base)
                            } else {
                                st.fail[base] = 0
                            }
                            st.lbMu.Unlock()
                            return err
                        }

                        // gRPC proxy via h2c reverse proxy (net/http) mounted in Fiber
                        if cr.cfg.GrpcProxy {
                            if st.cb != nil && st.cb.State() == gobreaker.StateOpen {
                                return c.Status(fiber.StatusServiceUnavailable).SendString("circuit open")
                            }
                            h := newH2cReverseProxy(base, cr.cfg.RewritePrefix)
                            if st.cb == nil {
                                st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                                err := adaptor.HTTPHandler(h)(c)
                                st.lbMu.Lock()
                                st.active[base]--
                                if err != nil {
                                    st.fail[base]++
                                    maybeEject(cr.cfg, st, base)
                                } else {
                                    st.fail[base] = 0
                                }
                                st.lbMu.Unlock()
                                return err
                            }
                            // run via breaker
                            _, err := st.cb.Execute(func() (interface{}, error) {
                                st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                                e := adaptor.HTTPHandler(h)(c)
                                st.lbMu.Lock()
                                st.active[base]--
                                if e != nil {
                                    st.fail[base]++
                                    maybeEject(cr.cfg, st, base)
                                } else {
                                    st.fail[base] = 0
                                }
                                st.lbMu.Unlock()
                                return nil, e
                            })
                            return err
                        }

                        // default http proxy
                        target := base + targetPath
                        if st.cb == nil {
                            st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                            err := proxy.Do(c, target)
                            st.lbMu.Lock()
                            st.active[base]--
                            if err != nil {
                                st.fail[base]++
                                maybeEject(cr.cfg, st, base)
                            } else {
                                st.fail[base] = 0
                            }
                            st.lbMu.Unlock()
                            return err
                        }
                        _, err := st.cb.Execute(func() (interface{}, error) {
                            st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                            e := proxy.Do(c, target)
                            st.lbMu.Lock()
                            st.active[base]--
                            if e != nil {
                                st.fail[base]++
                                maybeEject(cr.cfg, st, base)
                            } else {
                                st.fail[base] = 0
                            }
                            st.lbMu.Unlock()
                            return nil, e
                        })
                        return err
                    }
                    return c.SendStatus(fiber.StatusNotFound)
                })
            }
        }
    }

    // Health
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok", "service": "fiber-gateway"})
    })

    // Fallback routes when no config is provided
    if secrets.GetOrDefault("GW_CONFIG", "") == "" {
        attachDefaultRoutes(app)
    }

    // Regex route example (future use):
    // app.All("/api/(?i)public/.*", someHandler)

    // Start server
    port := getEnv("GW_PORT", "8090")
    log.Printf("fiber-gateway listening on :%s", port)
    if err := app.Listen(":" + port); err != nil {
        log.Fatalf("gateway error: %v", err)
    }
}

// helper functions moved to helpers.go, redis_helpers.go, outlier.go


