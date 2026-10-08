package verifier

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"
)

var (
	// ErrNullMX indicates the domain explicitly publishes a Null MX record (RFC 7505), declining all email.
	ErrNullMX = errors.New("domain publishes a null MX record and does not accept email")
	// ErrNoMailServer indicates neither MX records nor fallback A/AAAA records exist for the domain.
	ErrNoMailServer = errors.New("no MX or fallback A/AAAA records found for domain")
)

// DNSResolver handles DNS queries for mail exchange records.
type DNSResolver struct {
	resolver *net.Resolver
	timeout  time.Duration
}

// ResolverOption configures the DNSResolver.
type ResolverOption func(*DNSResolver)

// WithDNSServer configures the resolver to use a custom nameserver (e.g. "8.8.8.8:53" or "1.1.1.1:53").
func WithDNSServer(server string) ResolverOption {
	return func(r *DNSResolver) {
		if server == "" {
			return
		}
		if !strings.Contains(server, ":") {
			server = net.JoinHostPort(server, "53")
		}
		r.resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: r.timeout}
				return d.DialContext(ctx, "udp", server)
			},
		}
	}
}

// WithDNSTimeout sets the timeout for DNS queries.
func WithDNSTimeout(d time.Duration) ResolverOption {
	return func(r *DNSResolver) {
		if d > 0 {
			r.timeout = d
		}
	}
}

// NewDNSResolver creates a new DNSResolver with the given options.
func NewDNSResolver(opts ...ResolverOption) *DNSResolver {
	r := &DNSResolver{
		resolver: net.DefaultResolver,
		timeout:  5 * time.Second,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Lookup queries the domain for MX records, falling back to A/AAAA records per RFC 5321.
func (r *DNSResolver) Lookup(ctx context.Context, domain string) (DNSResult, error) {
	domain = strings.TrimSuffix(strings.TrimSpace(domain), ".")
	if domain == "" {
		return DNSResult{
			HasMX:  false,
			Reason: "domain cannot be empty",
		}, errors.New("domain cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// 1. Query MX Records
	mxList, err := r.resolver.LookupMX(ctx, domain)
	if err == nil && len(mxList) > 0 {
		// Check for Null MX record (RFC 7505: single MX with host "." and preference 0)
		if len(mxList) == 1 {
			host := strings.TrimSuffix(mxList[0].Host, ".")
			if host == "" || host == "." {
				return DNSResult{
					HasMX:  false,
					Reason: "domain publishes a Null MX record (RFC 7505) and rejects all email",
				}, ErrNullMX
			}
		}

		// Sort MX records by Preference ascending (lower number = higher priority)
		sort.Slice(mxList, func(i, j int) bool {
			return mxList[i].Pref < mxList[j].Pref
		})

		records := make([]MXRecord, 0, len(mxList))
		for _, mx := range mxList {
			cleanHost := strings.TrimSuffix(mx.Host, ".")
			records = append(records, MXRecord{
				Host: cleanHost,
				Pref: mx.Pref,
			})
		}

		return DNSResult{
			HasMX:     true,
			MXRecords: records,
		}, nil
	}

	// 2. RFC 5321 Section 5.1 Fallback: Query A/AAAA records if no MX is found
	ips, ipErr := r.resolver.LookupIP(ctx, "ip", domain)
	if ipErr == nil && len(ips) > 0 {
		// Domain itself acts as fallback MX with default preference 0
		fallbackRecord := MXRecord{
			Host: domain,
			Pref: 0,
		}
		return DNSResult{
			HasMX:     false,
			HasA:      true,
			MXRecords: []MXRecord{fallbackRecord},
			Reason:    "no MX records found; using domain A/AAAA record as fallback per RFC 5321",
		}, nil
	}

	// Neither MX nor A/AAAA records exist
	reason := fmt.Sprintf("no mail exchange found for domain %q", domain)
	if err != nil {
		reason = fmt.Sprintf("DNS lookup failed: %v", err)
	}

	return DNSResult{
		HasMX:  false,
		HasA:   false,
		Reason: reason,
	}, ErrNoMailServer
}

