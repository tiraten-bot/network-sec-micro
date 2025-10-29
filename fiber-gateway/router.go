package main

import (
    "context"
    "encoding/json"
    "net/http"
    "sync"
    "time"
    "strconv"

    adaptor "github.com/gofiber/adaptor/v2"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/proxy"
    "github.com/redis/go-redis/v9"
    "github.com/sony/gobreaker"
    "golang.org/x/time/rate"
)

// routeState holds per-route runtime state (rate limiters, breaker, LB counters)
type routeState struct {
    rlMu         sync.Mutex
    rlMap        map[string]*rate.Limiter
    cb           *gobreaker.CircuitBreaker
    lbMu         sync.Mutex
    lbIdx        int
    active       map[string]int
    fail         map[string]int
    ejectedUntil map[string]time.Time
}

// attachConfiguredRoutes installs a single catch-all handler that matches
// declarative routes in order and proxies accordingly.
func attachConfiguredRoutes(app *fiber.App, croutes []compiledRoute, rdb *redis.Client) {
    app.All("/*", MakeRouteHandler(croutes, rdb))
}

// MakeRouteHandler builds a catch-all handler for the provided compiled routes
func MakeRouteHandler(croutes []compiledRoute, rdb *redis.Client) fiber.Handler {
    states := make([]routeState, len(croutes))
    cache := newMemoryCache()

    // init circuit breakers and maps
    for i := range croutes {
        rc := croutes[i].cfg
        if rc.CircuitBreaker != nil && rc.CircuitBreaker.Enabled {
            st := &states[i]
            st.cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
                Name:     rc.Name,
                Interval: time.Duration(rc.CircuitBreaker.IntervalSec) * time.Second,
                Timeout:  time.Duration(rc.CircuitBreaker.TimeoutSec) * time.Second,
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

    return func(c *fiber.Ctx) error {
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
                    win := cr.cfg.RateLimit.WindowSec; if win <= 0 { win = 1 }
                    lim := cr.cfg.RateLimit.Limit; if lim <= 0 { lim = 100 }
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
                    tomorrow := time.Date(ts.Year(), ts.Month(), ts.Day()+1, 0, 0, 0, 0, time.UTC)
                    ttl := int(time.Until(tomorrow).Seconds())
                    rkey := "q:d:" + cr.cfg.Name + ":" + qkey + ":" + base
                    ok, err := redisQuota(c.Context(), rdb, rkey, cr.cfg.Quota.Daily, ttl)
                    if err != nil { return c.Status(fiber.StatusInternalServerError).SendString("quota err") }
                    if !ok { return c.Status(fiber.StatusTooManyRequests).SendString("daily quota exceeded") }
                }
            }

            // transforms
            for k, v := range cr.cfg.HeadersSet { c.Request().Header.Set(k, v) }
            for _, k := range cr.cfg.HeadersRemove { c.Request().Header.Del(k) }
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

            // load balancer pick
            if len(cr.cfg.Upstreams) > 0 {
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

            // Aggregation path (if configured)
            if len(cr.cfg.Aggregates) > 0 {
                // caching for aggregate
                if cr.cfg.Cache != nil && cr.cfg.Cache.Enabled && c.Method() == fiber.MethodGet {
                    key := "agg:" + path
                    if len(cr.cfg.Cache.VaryHeaders) > 0 {
                        for _, h := range cr.cfg.Cache.VaryHeaders { key += "|" + c.Get(h) }
                    }
                    if ce, ok := cache.get(key); ok {
                        if inm := c.Get("If-None-Match"); inm != "" && inm == ce.etag {
                            return c.SendStatus(fiber.StatusNotModified)
                        }
                        for hk, hv := range ce.headers { c.Set(hk, hv) }
                        c.Set("ETag", ce.etag)
                        c.Type("json")
                        return c.Status(fiber.StatusOK).Send(ce.payload)
                    }
                    res, status, headers, err := executeAggregate(c, cr)
                    if err != nil { return err }
                    body, _ := json.Marshal(res)
                    etag := strconv.FormatInt(int64(len(body)), 16) + ":" + fmtInt(time.Now().UnixNano())
                    ce := cacheEntry{expiresAt: time.Now().Add(time.Duration(cr.cfg.Cache.TtlSec)*time.Second), etag: etag, payload: body, headers: headers}
                    cache.set(key, ce)
                    for hk, hv := range headers { c.Set(hk, hv) }
                    c.Set("ETag", etag)
                    c.Type("json")
                    return c.Status(status).Send(body)
                }
                // without cache
                res, status, headers, err := executeAggregate(c, cr)
                if err != nil { return err }
                for hk, hv := range headers { c.Set(hk, hv) }
                c.Type("json")
                body, _ := json.Marshal(res)
                return c.Status(status).Send(body)
            }

            // websocket passthrough
            if cr.cfg.WebSocketPassthrough && isWebSocket(c) {
                if st.cb != nil && st.cb.State() == gobreaker.StateOpen {
                    return c.Status(fiber.StatusServiceUnavailable).SendString("circuit open")
                }
                target := toWS(base) + targetPath
                st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                err := proxy.Do(c, target)
                st.lbMu.Lock(); st.active[base]--; if err != nil { st.fail[base]++; maybeEject(cr.cfg, st, base) } else { st.fail[base] = 0 } st.lbMu.Unlock()
                return err
            }

            // gRPC h2c proxy
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
                _, err := st.cb.Execute(func() (interface{}, error) {
                    st.lbMu.Lock(); st.active[base]++; st.lbMu.Unlock()
                    e := adaptor.HTTPHandler(h)(c)
                    st.lbMu.Lock(); st.active[base]--; if e != nil { st.fail[base]++; maybeEject(cr.cfg, st, base) } else { st.fail[base] = 0 } st.lbMu.Unlock()
                    return nil, e
                })
                return err
            }

            // default HTTP proxy
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
    }
}

// executeAggregate performs fan-out calls and merges JSON responses
func executeAggregate(c *fiber.Ctx, cr compiledRoute) (map[string]interface{}, int, map[string]string, error) {
    policy := cr.cfg.AggregatePolicy
    overallTimeout := 5 * time.Second
    if policy != nil && policy.TimeoutS > 0 { overallTimeout = time.Duration(policy.TimeoutS) * time.Second }
    ctx, cancel := context.WithTimeout(c.Context(), overallTimeout)
    defer cancel()

    type result struct {
        name string
        body map[string]interface{}
        err  error
    }
    resCh := make(chan result, len(cr.cfg.Aggregates))
    for _, call := range cr.cfg.Aggregates {
        call := call
        go func() {
            method := call.Method
            if method == "" { method = http.MethodGet }
            req, _ := http.NewRequestWithContext(ctx, method, call.URL, nil)
            for hk, hv := range call.Headers { req.Header.Set(hk, hv) }
            httpClient := &http.Client{ Timeout: time.Duration(call.TimeoutS) * time.Second }
            if call.TimeoutS == 0 { httpClient.Timeout = overallTimeout }
            resp, err := httpClient.Do(req)
            if err != nil { resCh <- result{name: call.Name, err: err}; return }
            defer resp.Body.Close()
            var m map[string]interface{}
            if err := json.NewDecoder(resp.Body).Decode(&m); err != nil { resCh <- result{name: call.Name, err: err}; return }
            resCh <- result{name: call.Name, body: m, err: nil}
        }()
    }

    merged := map[string]interface{}{}
    successes := 0
    for i := 0; i < len(cr.cfg.Aggregates); i++ {
        select {
        case r := <-resCh:
            if r.err != nil {
                if policy != nil && policy.Partial != "ignore" { return nil, fiber.StatusBadGateway, nil, r.err }
                continue
            }
            merged[r.name] = r.body
            successes++
        case <-ctx.Done():
            if policy != nil && policy.Partial != "ignore" { return nil, fiber.StatusGatewayTimeout, nil, ctx.Err() }
        }
    }

    // Simple mapping on merged root
    if policy != nil {
        for oldk, newk := range policy.Rename {
            if v, ok := merged[oldk]; ok {
                merged[newk] = v
                delete(merged, oldk)
            }
        }
        for _, k := range policy.Omit { delete(merged, k) }
        if policy.ResponseRoot != "" {
            merged = map[string]interface{}{ policy.ResponseRoot: merged }
        }
    }

    headers := map[string]string{"Content-Type": "application/json"}
    if successes == 0 { return merged, fiber.StatusBadGateway, headers, nil }
    return merged, fiber.StatusOK, headers, nil
}

// attachDefaultRoutes installs simple path-based static proxy rules when no config is provided.
func attachDefaultRoutes(app *fiber.App) {
    app.All("/api/warrior/*", MakeDefaultHandler())
    app.All("/api/enemy/*", MakeDefaultHandler())
    app.All("/api/dragon/*", MakeDefaultHandler())
    app.All("/api/weapon/*", MakeDefaultHandler())
}

// MakeDefaultHandler proxies to static upstreams based on the first path segment
func MakeDefaultHandler() fiber.Handler {
    warriorUp := getEnv("UPSTREAM_WARRIOR", "http://localhost:8080")
    warriorCanaryUp := getEnv("UPSTREAM_WARRIOR_CANARY", "")
    warriorCanaryPct := getEnv("UPSTREAM_WARRIOR_CANARY_PERCENT", "")
    enemyUp := getEnv("UPSTREAM_ENEMY", "http://localhost:8083")
    dragonUp := getEnv("UPSTREAM_DRAGON", "http://localhost:8084")
    weaponUp := getEnv("UPSTREAM_WEAPON", "http://localhost:8081")
    return func(c *fiber.Ctx) error {
        path := c.OriginalURL()
        if hasPrefix(path, "/api/warrior/") {
            targetBase := warriorUp
            if warriorCanaryUp != "" && warriorCanaryPct != "" {
                rid := c.Get("X-Request-ID")
                if len(rid) > 0 {
                    last := rid[len(rid)-1]
                    pct := 10
                    if p, err := parsePercent(warriorCanaryPct); err == nil { pct = p }
                    if int(last)%100 < pct { targetBase = warriorCanaryUp }
                }
            }
            target := targetBase + path[len("/api/warrior"):]
            return proxy.Do(c, target)
        }
        if hasPrefix(path, "/api/enemy/") {
            target := enemyUp + path[len("/api/enemy"):]
            return proxy.Do(c, target)
        }
        if hasPrefix(path, "/api/dragon/") {
            target := dragonUp + path[len("/api/dragon"):]
            return proxy.Do(c, target)
        }
        if hasPrefix(path, "/api/weapon/") {
            target := weaponUp + path[len("/api/weapon"):]
            return proxy.Do(c, target)
        }
        return c.SendStatus(fiber.StatusNotFound)
    }
}


