package api

import (
	"fmt"
	"net"
	"net/url"
)

// validateLLMBaseURL guards against SSRF via the user-supplied LLM base_url.
// It requires an http(s) URL and always blocks link-local addresses (e.g. the
// cloud metadata endpoint 169.254.169.254). Private/loopback hosts are allowed
// only when allowPrivate is true (needed for local LLMs such as Ollama).
func validateLLMBaseURL(raw string, allowPrivate bool) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing host")
	}

	var ips []net.IP
	if ip := net.ParseIP(host); ip != nil {
		ips = []net.IP{ip}
	} else {
		resolved, err := net.LookupIP(host)
		if err != nil {
			return fmt.Errorf("could not resolve host")
		}
		ips = resolved
	}

	for _, ip := range ips {
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("link-local addresses are not allowed")
		}
		if !allowPrivate && (ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified()) {
			return fmt.Errorf("private/loopback addresses are not allowed")
		}
	}
	return nil
}
