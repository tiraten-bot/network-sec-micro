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
    "context"
    redis "github.com/redis/go-redis/v9"
    "strconv"
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
                // Redis client (optional)
                var rdb *redis.Client
                if addr := os.Getenv("REDIS_ADDR"); addr != "" {
                    rdb = redis.NewClient(&redis.Options{
                        Addr:     addr,
                        Password: os.Getenv("REDIS_PASSWORD"),
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
                            st.lbMu.Lock(); st.active[base]--; if err != nil { st.fail[base]++; maybeEject(cr.cfg, st, base) } else { st.fail[base] = 0 } st.lbMu.Unlock()
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
                                st.lbMu.Lock(); st.active[base]--; if err != nil { st.fail[base]++; maybeEject(cr.cfg, st, base) } else { st.fail[base] = 0 } st.lbMu.Unlock()
                                return err
                            }
                            // run via breaker
                            _, err := st.cb.Execute(func() (interface{}, error) {
                                st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                                e := adaptor.HTTPHandler(h)(c)
                                st.lbMu.Lock(); st.active[base]--; if e != nil { st.fail[base]++; maybeEject(cr.cfg, st, base) } else { st.fail[base] = 0 } st.lbMu.Unlock()
                                return nil, e
                            })
                            return err
                        }

                        // default http proxy
                        target := base + targetPath
                        if st.cb == nil {
                            st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                            err := proxy.Do(c, target)
                            st.lbMu.Lock(); st.active[base]--; if err != nil { st.fail[base]++; maybeEject(cr.cfg, st, base) } else { st.fail[base] = 0 } st.lbMu.Unlock()
                            return err
                        }
                        _, err := st.cb.Execute(func() (interface{}, error) {
                            st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                            e := proxy.Do(c, target)
                            st.lbMu.Lock(); st.active[base]--; if e != nil { st.fail[base]++; maybeEject(cr.cfg, st, base) } else { st.fail[base] = 0 } st.lbMu.Unlock()
                            return nil, e
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

// Redis helpers
func redisAllow(ctx context.Context, rdb *redis.Client, key string, limit int, windowSec int) (bool, error) {
    // INCR and set EXPIRE if first; allow if value <= limit
    n, err := rdb.Incr(ctx, key).Result()
    if err != nil { return false, err }
    if n == 1 { _ = rdb.Expire(ctx, key, time.Duration(windowSec)*time.Second).Err() }
    return n <= int64(limit), nil
}

func redisQuota(ctx context.Context, rdb *redis.Client, key string, limit int, ttlSec int) (bool, error) {
    n, err := rdb.Incr(ctx, key).Result()
    if err != nil { return false, err }
    if n == 1 { _ = rdb.Expire(ctx, key, time.Duration(ttlSec)*time.Second).Err() }
    return n <= int64(limit), nil
}

func fmtInt(v int64) string { return strconv.FormatInt(v, 10) }

func maybeEject(rc RouteConfig, st *routeState, upstream string) {
    if rc.OutlierDetection == nil || !rc.OutlierDetection.Enabled { return }
    thr := rc.OutlierDetection.FailureThreshold
    if thr <= 0 { thr = 5 }
    if st.fail[upstream] >= thr {
        dur := time.Duration(rc.OutlierDetection.EjectDurationSec) * time.Second
        if dur <= 0 { dur = 30 * time.Second }
        st.ejectedUntil[upstream] = time.Now().Add(dur)
        st.fail[upstream] = 0
    }
}


