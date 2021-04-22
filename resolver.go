package discovery

import "net/url"

// Resolver specifies how to request for service discovery
type Resolver interface {
	// Resolve tries to resolve a service lookup query of the given service,
	Resolve(portName, serviceName string) (*url.URL, error)
}
