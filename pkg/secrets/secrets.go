package secrets

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	vault "github.com/hashicorp/vault/api"
)

// ErrNotFound is returned when a secret could not be located in any backend.
var ErrNotFound = errors.New("secret not found")

var (
	clientMu sync.Mutex
	client   *vault.Client

	cacheMu     sync.RWMutex
	cache       map[string]string
	cacheLoaded bool
)

// Get retrieves a secret value by key. It first checks environment variables,
// then falls back to Vault when Vault integration is enabled. Keys are treated
// as case-insensitive.
func Get(key string) (string, error) {
	normalized := normalizeKey(key)

	if value := strings.TrimSpace(os.Getenv(normalized)); value != "" {
		return value, nil
	}

	if !vaultEnabled() {
		return "", ErrNotFound
	}

	if err := ensureVaultClient(); err != nil {
		return "", fmt.Errorf("vault client init failed: %w", err)
	}

	if err := ensureCache(); err != nil {
		return "", err
	}

	cacheLock.RLock()
	defer cacheLock.RUnlock()

	if value, ok := cache[normalized]; ok && value != "" {
		return value, nil
	}

	return "", ErrNotFound
}

// GetOrDefault returns the secret value if present, otherwise the provided default.
func GetOrDefault(key, defaultValue string) string {
	if value, err := Get(key); err == nil && value != "" {
		return value
	}
	return defaultValue
}

// GetList splits a secret on the provided separator and returns a slice. When
// the secret is missing, the default slice is returned.
func GetList(key, separator string, defaultValue []string) []string {
	value, err := Get(key)
	if err != nil || strings.TrimSpace(value) == "" {
		return defaultValue
	}

	parts := strings.Split(value, separator)
	var cleaned []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	if len(cleaned) == 0 {
		return defaultValue
	}
	return cleaned
}

// MustGet behaves like Get but panics when the secret cannot be retrieved.
func MustGet(key string) string {
	value, err := Get(key)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve secret %q: %v", key, err))
	}
	return value
}

func vaultEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("VAULT_ENABLED"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func ensureVaultClient() error {
	clientMu.Lock()
	defer clientMu.Unlock()

	if client != nil {
		return nil
	}

	config := vault.DefaultConfig()

	if addr := strings.TrimSpace(os.Getenv("VAULT_ADDR")); addr != "" {
		config.Address = addr
	}

	if err := configureVaultTLS(config); err != nil {
		return fmt.Errorf("configure TLS: %w", err)
	}

	cli, err := vault.NewClient(config)
	if err != nil {
		return fmt.Errorf("create vault client: %w", err)
	}

	if ns := strings.TrimSpace(os.Getenv("VAULT_NAMESPACE")); ns != "" {
		cli.SetNamespace(ns)
	}

	if err := authenticateVaultClient(cli); err != nil {
		return err
	}

	client = cli
	return nil
}

func ensureCache() error {
	cacheMu.RLock()
	if cacheLoaded {
		cacheMu.RUnlock()
		return nil
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cacheLoaded {
		return nil
	}

	data, err := loadVaultSecrets()
	if err != nil {
		return err
	}

	cache = data
	cacheLoaded = true
	return nil
}

func loadVaultSecrets() (map[string]string, error) {
	path := strings.TrimSpace(os.Getenv("VAULT_SECRET_PATH"))
	if path == "" {
		path = "secret/data/network-sec-micro"
	}

	secret, err := client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("read vault path %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return map[string]string{}, nil
	}

	rawData := secret.Data
	// Handle KV v2 payload
	if dataField, ok := rawData["data"]; ok {
		if inner, ok := dataField.(map[string]interface{}); ok {
			rawData = inner
		}
	}

	resolved := make(map[string]string, len(rawData))
	for k, v := range rawData {
		switch typed := v.(type) {
		case string:
			resolved[normalizeKey(k)] = typed
		case fmt.Stringer:
			resolved[normalizeKey(k)] = typed.String()
		default:
			resolved[normalizeKey(k)] = fmt.Sprintf("%v", typed)
		}
	}

	return resolved, nil
}

func configureVaultTLS(config *vault.Config) error {
	tlsConfig := &vault.TLSConfig{}

	if caPath := strings.TrimSpace(os.Getenv("VAULT_CACERT")); caPath != "" {
		tlsConfig.CACert = caPath
	}

	if caDir := strings.TrimSpace(os.Getenv("VAULT_CAPATH")); caDir != "" {
		tlsConfig.CAPath = caDir
	}

	if cert := strings.TrimSpace(os.Getenv("VAULT_CLIENT_CERT")); cert != "" {
		tlsConfig.ClientCert = cert
	}

	if key := strings.TrimSpace(os.Getenv("VAULT_CLIENT_KEY")); key != "" {
		tlsConfig.ClientKey = key
	}

	if skip := strings.TrimSpace(os.Getenv("VAULT_SKIP_VERIFY")); skip != "" {
		switch strings.ToLower(skip) {
		case "1", "true", "yes", "on":
			tlsConfig.Insecure = true
		}
	}

	if tlsConfig.CACert == "" && tlsConfig.CAPath == "" && !tlsConfig.Insecure {
		return nil
	}

	return config.ConfigureTLS(tlsConfig)
}

func authenticateVaultClient(cli *vault.Client) error {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("VAULT_AUTH_METHOD"))) {
	case "approle":
		roleID := strings.TrimSpace(os.Getenv("VAULT_ROLE_ID"))
		secretID := strings.TrimSpace(os.Getenv("VAULT_SECRET_ID"))
		if roleID == "" || secretID == "" {
			return errors.New("approle authentication requested but VAULT_ROLE_ID/VAULT_SECRET_ID missing")
		}
		secret, err := cli.Logical().Write("auth/approle/login", map[string]interface{}{
			"role_id":   roleID,
			"secret_id": secretID,
		})
		if err != nil {
			return fmt.Errorf("vault approle login failed: %w", err)
		}
		if secret == nil || secret.Auth == nil || secret.Auth.ClientToken == "" {
			return errors.New("vault approle login returned empty token")
		}
		cli.SetToken(secret.Auth.ClientToken)
	default:
		token := strings.TrimSpace(os.Getenv("VAULT_TOKEN"))
		if token == "" {
			if tokenFile := strings.TrimSpace(os.Getenv("VAULT_TOKEN_FILE")); tokenFile != "" {
				content, err := os.ReadFile(filepath.Clean(tokenFile))
				if err != nil {
					return fmt.Errorf("read vault token file: %w", err)
				}
				token = strings.TrimSpace(string(content))
			}
		}

		if token == "" {
			return errors.New("vault token not provided")
		}

		cli.SetToken(token)
	}

	return nil
}

func normalizeKey(key string) string {
	return strings.ToUpper(strings.TrimSpace(key))
}

// Refresh flushes the cached secrets forcing the next lookup to hit Vault again.
func Refresh() error {
	if !vaultEnabled() {
		return nil
	}

	if err := ensureVaultClient(); err != nil {
		return err
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	data, err := loadVaultSecrets()
	if err != nil {
		return err
	}

	cache = data
	cacheLoaded = true

	return nil
}

// WaitForVault blocks until Vault responds successfully or the timeout is reached.
// This can be used in entrypoints to ensure the secrets backend is ready.
func WaitForVault(timeout time.Duration) error {
	if !vaultEnabled() {
		return nil
	}

	start := time.Now()
	for {
		err := ensureVaultClient()
		if err == nil {
			return nil
		}

		if time.Since(start) > timeout {
			return err
		}

		time.Sleep(2 * time.Second)
	}
}
