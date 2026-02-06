// Package security provides security utilities for the UGC API.
package security

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
)

// Common errors for URL validation.
var (
	ErrInvalidURL       = errors.New("invalid URL")
	ErrHTTPSRequired    = errors.New("URL must use HTTPS")
	ErrHostNotAllowed   = errors.New("host not in allowlist")
	ErrEmptyURL         = errors.New("URL is empty")
	ErrPrivateIPBlocked = errors.New("private IP addresses are not allowed")
	ErrDNSLookupFailed  = errors.New("DNS lookup failed for host")
)

// URLValidator validates URLs against a host allowlist to prevent SSRF attacks.
type URLValidator struct {
	mu           sync.RWMutex
	allowedHosts map[string]bool
}

// NewURLValidator creates a new URLValidator with the given allowed hosts.
// If no hosts are provided, defaultAllowedHosts() will be used.
func NewURLValidator(allowedHosts []string) *URLValidator {
	hosts := allowedHosts
	if len(hosts) == 0 {
		hosts = defaultAllowedHosts()
	}

	allowedMap := make(map[string]bool, len(hosts))
	for _, host := range hosts {
		// Normalize host to lowercase
		allowedMap[strings.ToLower(strings.TrimSpace(host))] = true
	}

	return &URLValidator{
		allowedHosts: allowedMap,
	}
}

// defaultAllowedHosts returns the default list of allowed hosts for media URLs.
// Returns a fresh copy each time to prevent external mutation.
func defaultAllowedHosts() []string {
	return []string{
		// Suno - allow all subdomains
		"suno.ai",
		"suno.com",
		"audiopipe.suno.ai",
		"cdn1.suno.ai",
		"cdn2.suno.ai",
		// KIE - allow all subdomains
		"kie.ai",
		"cdn.kie.ai",
		"storage.kie.ai",
		"musicfile.kie.ai",
		// AWS S3 (commonly used by Suno) — specific subdomains only
		"s3.amazonaws.com",
		"s3.us-east-1.amazonaws.com",
		"s3.us-west-2.amazonaws.com",
		// NanoBanana image hosts
		"nanobananastorage.blob.core.windows.net",
		"aiquickdraw.com",
	}
}

// ValidateURL validates that a URL is safe to fetch.
// It checks:
// 1. URL is not empty
// 2. URL is valid and parseable
// 3. URL uses HTTPS scheme
// 4. Host is in the allowlist
// 5. Host does not resolve to private IP
func (v *URLValidator) ValidateURL(rawURL string) error {
	if rawURL == "" {
		return ErrEmptyURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	// Require HTTPS for security
	if parsed.Scheme != "https" {
		return ErrHTTPSRequired
	}

	host := strings.ToLower(parsed.Hostname())

	// Hold read lock for allowlist check only (DNS check runs independently after)
	v.mu.RLock()
	allowed := v.isAllowedHostLocked(host)
	v.mu.RUnlock()

	if !allowed {
		return ErrHostNotAllowed
	}

	// Block private/internal IPs to prevent SSRF (fail-closed on DNS error)
	if err := checkNotPrivateIP(host); err != nil {
		return err
	}

	return nil
}

// checkNotPrivateIP resolves the host and verifies none of the IPs are private/internal.
// Fails closed: returns error if DNS resolution fails (prevents bypass via DNS failure).
func checkNotPrivateIP(host string) error {
	ips, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrDNSLookupFailed, host)
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return ErrPrivateIPBlocked
		}
	}

	return nil
}

// isAllowedHostLocked checks if the given host is in the allowlist.
// Caller MUST hold v.mu.RLock().
func (v *URLValidator) isAllowedHostLocked(host string) bool {
	// Direct match
	if v.allowedHosts[host] {
		return true
	}

	// Check if it's a subdomain of an allowed host
	for allowedHost := range v.allowedHosts {
		if strings.HasSuffix(host, "."+allowedHost) {
			return true
		}
	}

	return false
}

// AddHost adds a host to the allowlist.
func (v *URLValidator) AddHost(host string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.allowedHosts[strings.ToLower(strings.TrimSpace(host))] = true
}

// RemoveHost removes a host from the allowlist.
func (v *URLValidator) RemoveHost(host string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.allowedHosts, strings.ToLower(strings.TrimSpace(host)))
}

// AllowedHosts returns a copy of the current allowlist.
func (v *URLValidator) AllowedHosts() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	hosts := make([]string, 0, len(v.allowedHosts))
	for host := range v.allowedHosts {
		hosts = append(hosts, host)
	}
	return hosts
}
