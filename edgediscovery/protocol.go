package edgediscovery

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	protocolRecord          = "protocol-v2.argotunnel.com"
	protocolLookupTimeout   = 10 * time.Second
	cloudflareDOTServerName = "cloudflare-dns.com"
	cloudflareDOTServerAddr = "1.1.1.1:853"
)

var (
	errNoProtocolRecord = fmt.Errorf("No TXT record found for %s to determine connection protocol", protocolRecord)
)

type PercentageFetcher func() (ProtocolPercents, error)

// ProtocolPercent represents a single Protocol Percentage combination.
type ProtocolPercent struct {
	Protocol   string `json:"protocol"`
	Percentage int32  `json:"percentage"`
}

// ProtocolPercents represents the preferred distribution ratio of protocols when protocol isn't specified.
type ProtocolPercents []ProtocolPercent

// GetPercentage returns the threshold percentage of a single protocol.
func (p ProtocolPercents) GetPercentage(protocol string) int32 {
	for _, protocolPercent := range p {
		if strings.ToLower(protocolPercent.Protocol) == strings.ToLower(protocol) {
			return protocolPercent.Percentage
		}
	}
	return 0
}

// ProtocolPercentage returns the ratio of protocols and a specification ratio for their selection.
func ProtocolPercentage() (ProtocolPercents, error) {
	ctx, cancel := context.WithTimeout(context.Background(), protocolLookupTimeout)
	defer cancel()

	records, err := newCloudflareDOTResolver().LookupTXT(ctx, protocolRecord)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errNoProtocolRecord
	}

	var protocolsWithPercent ProtocolPercents
	err = json.Unmarshal([]byte(records[0]), &protocolsWithPercent)
	return protocolsWithPercent, err
}

func newCloudflareDOTResolver() *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _ string, _ string) (net.Conn, error) {
			var dialer net.Dialer
			conn, err := dialer.DialContext(ctx, "tcp", cloudflareDOTServerAddr)
			if err != nil {
				return nil, err
			}
			tlsConfig := &tls.Config{ServerName: cloudflareDOTServerName}
			return tls.Client(conn, tlsConfig), nil
		},
	}
}
