package discovery

import (
	"net/url"
)

// schemeDecorator replaces url scheme if mapped
type schemeDecorator struct {
	innerResolver  Resolver
	portNameScheme map[string]string
}

func NewSchemeDecorator(innerResolver Resolver, portNameScheme map[string]string) Resolver {
	return &schemeDecorator{innerResolver: innerResolver, portNameScheme: portNameScheme}
}

func (s schemeDecorator) Resolve(portName, serviceName string) (*url.URL, error) {
	if scheme, ok := s.portNameScheme[portName]; ok {
		serviceUrl, err := s.innerResolver.Resolve(portName, serviceName)
		if serviceUrl != nil {
			serviceUrl.Scheme = scheme
		}
		return serviceUrl, err
	}
	return s.innerResolver.Resolve(portName, serviceName)
}
