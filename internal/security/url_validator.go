// Package security provides security utilities for the UGC API.
package security

import (
	"errors"
	"net/url"
	"strings"
)

// Common errors for URL validation.
var (
	ErrInvalidURL       = errors.New("invalid URL")
	ErrHTTPSRequired    = errors.New("URL must use HTTPS")
	ErrHostNotAllowed   = errors.New("host not in allowlist")
	ErrEmptyURL         = errors.New("URL is empty")
	ErrPrivateIPBlocked = errors.New("private IP addresses are not allowed")
)

// URLValidator validates URLs against a host allowlist to prevent SSRF attacks.
type URLValidator struct {
	allowedHosts map[string]bool
}

// NewURLValidator creates a new URLValidator with the given allowed hosts.
// If no hosts are provided, DefaultAllowedHosts will be used.
func NewURLValidator(allowedHosts []string) *URLValidator {
	hosts := allowedHosts
	if len(hosts) == 0 {
		hosts = DefaultAllowedHosts
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

// DefaultAllowedHosts contains the default list of allowed hosts for media URLs.
var DefaultAllowedHosts = []string{
	// Suno CDN
	"cdn1.suno.ai",
	"cdn2.suno.ai",
	// KIE CDN
	"cdn.kie.ai",
	"storage.kie.ai",
	// Common CDN providers (add as needed)
	"cdn.suno.com",
}

// ValidateURL validates that a URL is safe to fetch.
// It checks:
// 1. URL is not empty
// 2. URL is valid and parseable
// 3. URL uses HTTPS scheme
// 4. Host is in the allowlist
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

	// Check if host is in allowlist
	host := strings.ToLower(parsed.Hostname())
	if !v.isAllowedHost(host) {
		return ErrHostNotAllowed
	}

	return nil
}

// isAllowedHost checks if the given host is in the allowlist.
// It also checks for subdomain matches (e.g., "sub.cdn.kie.ai" matches "cdn.kie.ai").
func (v *URLValidator) isAllowedHost(host string) bool {
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
	v.allowedHosts[strings.ToLower(strings.TrimSpace(host))] = true
}

// RemoveHost removes a host from the allowlist.
func (v *URLValidator) RemoveHost(host string) {
	delete(v.allowedHosts, strings.ToLower(strings.TrimSpace(host)))
}

// AllowedHosts returns a copy of the current allowlist.
func (v *URLValidator) AllowedHosts() []string {
	hosts := make([]string, 0, len(v.allowedHosts))
	for host := range v.allowedHosts {
		hosts = append(hosts, host)
	}
	return hosts
}
