package dnsresolver

import (
	"context"
	"fmt"
	"net"
)

var cloudflareResolverAddrs = []string{
	"1.1.1.1:53",
	"1.0.0.1:53",
	"[2606:4700:4700::1111]:53",
	"[2606:4700:4700::1001]:53",
}

// NewCloudflareResolver returns a pure-Go resolver that always queries Cloudflare DNS directly.
//
// On Android this avoids the stdlib fallback to loopback resolvers such as 127.0.0.1:53 or [::1]:53,
// while still letting VPN/TUN DNS hijack rules intercept ordinary port-53 traffic if the environment wants to.
func NewCloudflareResolver() *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network string, _ string) (net.Conn, error) {
			var dialer net.Dialer
			var lastErr error
			for _, server := range cloudflareResolverAddrs {
				conn, err := dialer.DialContext(ctx, network, server)
				if err == nil {
					return conn, nil
				}
				lastErr = err
			}
			if lastErr == nil {
				lastErr = fmt.Errorf("no Cloudflare DNS resolver addresses configured")
			}
			return nil, lastErr
		},
	}
}
