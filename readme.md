# Discovery

Package discovery provides features that allows discovering services by names.

`go get github.com/Confialink/wallet-pkg-discovery/v2`


```go
usersServiceApiEnpointUrl, err := resolver.Resolve("public-api-port", "users")
// ...
someServiceRpcEndpointUrl, err := resolver.Resolve("rpc", "some-service")
// ... 
``` 

The package comes with several implementations described below. 

## DNS service discovery

DNS service discovery resolver aimed to perform SRV records lookup over DNS service provider.

```go
    import (
     "fmt"    
     discovery "github.com/Confialink/wallet-pkg-discovery/v2"
    )
    
    // see net.LookupSRV
    dnsResolver := discovery.NewDNSResolver(&discovery.NetSRVResolver{}, "tcp")
    
    serviceUrl, err := dnsResolver.Resolve("api-port-name", "service")
    // ...
    // prints something like this: tcp://_api-port-name._tcp.service.:12345
    fmt.Println(serviceUrl.String())
    
```

## Environment variable service discovery

Environment variable service discovery resolver uses OS environment variables in order to compose service url.

The resolver looks for variables which names are:
* `{SVCNAME}_SERVICE_HOST` - where {SVCNAME} is a name of a service.
* `{SVCNAME}_SERVICE_PORT_{PORTNAME}` - where {SVCNAME} is a name of a service, {PORTNAME} is a name of a port.

For example if service name is *users* and port name is *api* then environment variables would look like:
* `USERS_SERVICE_HOST=10.0.4.12`
* `USERS_SERVICE_PORT_API=8080`

```go
    import (
     "fmt"
     "net"     
     discovery "github.com/Confialink/wallet-pkg-discovery/v2"
    )
    
    envResolver := discovery.NewEnvResolver()
    
    serviceUrl, err := envResolver.Resolve("api", "users")
    serviceUrl.Scheme = "http"
    // ...
    // prints something like this: http://10.0.4.12:8080
    fmt.Println(serviceUrl.String())    
```

Note that while forming environment variable name the resolver removes all characters that cannot be used in a variable name.
Allowed characters must match this regular expression `[^a-zA-Z0-9_]+`.
Dashes are replaced with an underscores.
E.g. `private-port` becomes `PRIVATE_PORT`.

## Fallback service discovery

Fallback service discovery combines another resolvers so that if one fails another one is used until service is found or 
all resolvers checked.

```go
// ...
fallbackResolver := discovery.NewFallbackResolver(dnsResolver, envResolver)
// In this case, dnsResolver will be called first, and if it returns an error, then envResolver will be used.
serviceUrl, err := fallbackResolver.Resolve("api", "users")
```

## Scheme decorator

Scheme decorator is a special resolver that replaces URL scheme by mapping it to a port name.
Consider it as a helper that helps to reduce code duplication.

```go
 portNameSchemeMapping := map[string][string]{
    "api": "http",
    "private-api": "https",
 }
 
  apiUrl, _ := realResolver.Resolve("api", "service")
  // Let's say it prints: tcp://example.com
  fmt.Println(apiUrl.String()) 

 decorator := discovery.NewSchemeDecorator(realResolver, portNameSchemeMapping) 
 apiUrlDecorated, _ := descorator.Resolve("api", "service")
 // should print : http://example.com
 fmt.Println(apiUrlDecorated.String())
```

## Custom service discovery resolvers

You may want to implement your own service discovery logic. In order to do so simply implement `discovery.Resolver` interface.
```go
// Resolver specifies how to request for service discovery
type Resolver interface {
	// Resolve tries to resolve a service lookup query of the given service,
	Resolve(portName, serviceName string) (*url.URL, error)
}
```

For example resolver which always returns the same value:
```go
type MyFixedValueResolver struct {
    url.URL
}

func(m *MyFixedValueResolver) Resolve(portName, serviceName string) (*url.URL, error) {
	return &m.URL
}
// ...
fixedUrlResolver := MyFixedValueResolver{url.URL{Scheme: "http", Host: "1.2.3.4:5678"}}
// ...
serviceUrl, _ := fixedUrlResolver.Resolve("does-not", "matter")

// prints: "http://1.2.3.4:5678"
fmt.Println(serviceUrl.String())
```