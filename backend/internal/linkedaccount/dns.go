package linkedaccount

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func ExtractDomain(rawURL string) (string, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing URL: %w", err)
	}
	host := parsed.Hostname()
	if host == "" {
		return "", fmt.Errorf("no hostname found in URL")
	}
	return host, nil
}

func ValidatePublicDomain(domain string) error {
	if domain == "localhost" || strings.HasSuffix(domain, ".local") || strings.HasSuffix(domain, ".internal") {
		return fmt.Errorf("private domains are not allowed")
	}

	ip := net.ParseIP(domain)
	if ip != nil {
		return fmt.Errorf("IP addresses are not allowed, use a domain name")
	}

	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return fmt.Errorf("domain must have at least two parts")
	}

	return nil
}

// ValidateDomainResolution resolves the domain's A/AAAA records and rejects
// any that point to loopback, private, link-local, or unspecified addresses.
// This prevents SSRF attacks where a publicly-named domain resolves to an
// internal network address.
func ValidateDomainResolution(domain string) error {
	addrs, err := net.LookupHost(domain)
	if err != nil {
		return fmt.Errorf("DNS resolution failed for %s: %w", domain, err)
	}

	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			return fmt.Errorf("could not parse resolved address %s", addr)
		}

		if ip.IsLoopback() {
			return fmt.Errorf("domain resolves to a loopback address (%s)", addr)
		}
		if ip.IsPrivate() {
			return fmt.Errorf("domain resolves to a private address (%s)", addr)
		}
		if ip.IsLinkLocalUnicast() {
			return fmt.Errorf("domain resolves to a link-local unicast address (%s)", addr)
		}
		if ip.IsLinkLocalMulticast() {
			return fmt.Errorf("domain resolves to a link-local multicast address (%s)", addr)
		}
		if ip.IsUnspecified() {
			return fmt.Errorf("domain resolves to an unspecified address (%s)", addr)
		}
	}

	return nil
}

func VerifyDNSTXT(domain, expectedToken string) (bool, error) {
	records, err := net.LookupTXT(domain)
	if err != nil {
		return false, fmt.Errorf("DNS lookup failed for %s: %w", domain, err)
	}
	for _, record := range records {
		if strings.TrimSpace(record) == expectedToken {
			return true, nil
		}
	}
	return false, nil
}
