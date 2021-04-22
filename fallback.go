package discovery

import (
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

const maxCalls = 128

type fallbackResolver struct {
	resolvers []Resolver
}

func NewFallbackResolver(resolvers ...Resolver) Resolver {
	if len(resolvers) == 0 {
		panic("at least one resolver is expected")
	}
	return &fallbackResolver{resolvers: resolvers}
}

func (f *fallbackResolver) Resolve(portName, serviceName string) (*url.URL, error) {
	ignoredErrors := make([]string, 0, len(f.resolvers))
	c := 0
	for _, r := range f.resolvers {
		c++
		// in case if there is infinite loop (e.g. same fallback resolver is passed)
		if c > maxCalls {
			return nil, errors.New("failed to call Resolve: maximum number of calls exceeded")
		}
		serviceUrl, err := r.Resolve(portName, serviceName)
		if err != nil {
			ignoredErrors = append(ignoredErrors, err.Error())
			continue
		}
		return serviceUrl, nil
	}
	m := fmt.Sprintf(
		"failed to resolve \"%s\" with the given port \"%s\" URL: all resolvers are failed with errors: \n%s",
		serviceName,
		portName,
		strings.Join(ignoredErrors, "\n"),
	)
	return nil, errors.New(m)
}
