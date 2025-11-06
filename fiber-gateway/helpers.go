package main

import (
    "crypto/tls"
    "net"
    "net/http"
    "net/http/httputil"
    "strconv"
    "sync"
    "time"

    "network-sec-micro/pkg/secrets"

    "github.com/gofiber/fiber/v2"
    "golang.org/x/net/http2"
)

func getEnv(key, def string) string {
    return secrets.GetOrDefault(key, def)
}

func parsePercent(s string) (int, error) {
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
    if v < 0 { v = 0 }
    if v > 100 { v = 100 }
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
        targetPath := r.URL.Path
        if stripPrefix != "" && hasPrefix(targetPath, stripPrefix) {
            targetPath = targetPath[len(stripPrefix):]
            if targetPath == "" || targetPath[0] != '/' {
                targetPath = "/" + targetPath
            }
        }
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
    }
    rp := &httputil.ReverseProxy{Director: director}
    rp.Transport = &http2.Transport{
        AllowHTTP: true,
        DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
            return net.Dial(network, addr)
        },
    }
    return rp
}

func fmtInt(v int64) string { return strconv.FormatInt(v, 10) }

// simple in-memory cache (best-effort)
type cacheEntry struct {
    expiresAt time.Time
    etag      string
    payload   []byte
    headers   map[string]string
}

type memoryCache struct {
    mu sync.Mutex
    m  map[string]cacheEntry
}

func newMemoryCache() *memoryCache { return &memoryCache{m: make(map[string]cacheEntry)} }

func (c *memoryCache) get(key string) (cacheEntry, bool) {
    c.mu.Lock(); defer c.mu.Unlock()
    e, ok := c.m[key]
    if !ok || time.Now().After(e.expiresAt) { return cacheEntry{}, false }
    return e, true
}

func (c *memoryCache) set(key string, e cacheEntry) {
    c.mu.Lock(); c.m[key] = e; c.mu.Unlock()
}


