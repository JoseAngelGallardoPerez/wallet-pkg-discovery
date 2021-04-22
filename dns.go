package discovery

import (
	"fmt"
	"net"
	"net/url"

	"github.com/pkg/errors"
)

// SRVResolver determines DNS SRV record lookup contract
type SRVResolver interface {
	// LookupSRV tries to resolve an SRV query of the given service,
	// protocol, and domain name. The proto is "tcp" or "udp".
	// The returned records should be sorted by priority and randomized
	// by weight within a priority.
	LookupSRV(service, proto, name string) (cname string, addrs []*net.SRV, err error)
}

// NetSRVResolver uses default net.LookupSRV implementation
type NetSRVResolver struct{}

func (*NetSRVResolver) LookupSRV(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	return net.LookupSRV(service, proto, name)
}

func NewDNSResolver(srvResolver SRVResolver, proto string) Resolver {
	if proto != "tcp" && proto != "udp" {
		panic(fmt.Sprintf(`invalid argument "proto" is given - expected "tcp" or "udp", got "%s"`, proto))
	}
	return &defaultDNSResolver{
		srvResolver: srvResolver,
		proto:       proto,
	}
}

type defaultDNSResolver struct {
	srvResolver SRVResolver
	proto       string
}

func (d *defaultDNSResolver) Resolve(portName, serviceName string) (*url.URL, error) {
	_, addrs, err := d.srvResolver.LookupSRV(portName, d.proto, serviceName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to lookup SRV DNS record")
	}

	if len(addrs) == 0 {
		return nil, errors.New("record is not found")
	}

	addr := addrs[0]
	return &url.URL{
		Scheme: d.proto,
		Host:   fmt.Sprintf("%s:%d", addr.Target, addr.Port),
	}, nil
}
