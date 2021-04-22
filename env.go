package discovery

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const envVarNameValidCharactersStr = `[^a-zA-Z0-9_]+`

var envVarNameValidCharactersRegExp = regexp.MustCompile(envVarNameValidCharactersStr)

type envResolver struct{}

func NewEnvResolver() Resolver {
	return &envResolver{}
}

func (*envResolver) Resolve(portName, serviceName string) (*url.URL, error) {

	transformedPortName := envNormalizeName(portName)
	transformedServiceName := envNormalizeName(serviceName)

	envVarPortName := fmt.Sprintf(
		"%s_SERVICE_PORT_%s",
		transformedServiceName,
		transformedPortName,
	)
	envVarServiceName := fmt.Sprintf(
		"%s_SERVICE_HOST",
		transformedServiceName,
	)

	host := os.Getenv(envVarServiceName)
	if host == "" {
		msg := fmt.Sprintf(
			`failed to find service(%s) host: "%s" environment variable is not set`,
			serviceName,
			envVarServiceName,
		)
		return nil, errors.New(msg)
	}

	port := os.Getenv(envVarPortName)
	if port == "" {
		msg := fmt.Sprintf(
			`failed to find service(%s) port(%s): "%s" environment variable is not set`,
			serviceName,
			portName,
			envVarPortName,
		)
		return nil, errors.New(msg)
	}

	return &url.URL{
		Host: fmt.Sprintf("%s:%s", host, port),
	}, nil
}

func envNormalizeName(name string) string {
	name = strings.Replace(name, "-", "_", -1)
	name = strings.ToUpper(name)
	name = envVarNameValidCharactersRegExp.ReplaceAllString(name, "")
	return name
}
