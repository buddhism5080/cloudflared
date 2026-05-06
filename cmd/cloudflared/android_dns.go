//go:build android

package main

import (
	"net"

	"github.com/cloudflare/cloudflared/dnsresolver"
)

func init() {
	net.DefaultResolver = dnsresolver.NewCloudflareResolver()
}
