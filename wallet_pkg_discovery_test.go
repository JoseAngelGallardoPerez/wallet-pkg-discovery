package discovery_test

import (
	discovery "github.com/Confialink/wallet-pkg-discovery/v2"
	"github.com/Confialink/wallet-pkg-discovery/v2/mock_discovery"
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"net/url"
	"os"
)

var _ = Describe("WalletPkgDiscovery", func() {
	Context("DNS resolver", func() {
		When("SRV resolver returns one or more addresses", func() {
			It("should take first address and form URL", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				mockSrvResolver := mock_discovery.NewMockSRVResolver(ctrl)
				mockSrvResolver.
					EXPECT().
					LookupSRV("port-name", "tcp", "service-name").
					Return(
						"_port-name._tcp.domain.",
						[]*net.SRV{
							{
								Target:   "domain.svc.local.",
								Port:     12308,
								Priority: 0,
								Weight:   100,
							},
							{
								Target:   "domain.svc.local.",
								Port:     9999,
								Priority: 0,
								Weight:   100,
							},
						},
						nil,
					)

				resolver := discovery.NewDNSResolver(mockSrvResolver, "tcp")
				serviceURL, err := resolver.Resolve("port-name", "service-name")

				Expect(err).ShouldNot(HaveOccurred())
				Expect(serviceURL.String()).To(Equal("tcp://domain.svc.local.:12308"))
			})
		})

		When("SRV resolver returns empty addresses slice", func() {
			It("should return error", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				mockSrvResolver := mock_discovery.NewMockSRVResolver(ctrl)
				mockSrvResolver.
					EXPECT().
					LookupSRV("unknown-port", "tcp", "service-name").
					Return("", []*net.SRV{}, nil)

				resolver := discovery.NewDNSResolver(mockSrvResolver, "tcp")
				serviceURL, err := resolver.Resolve("unknown-port", "service-name")

				Expect(err).Should(HaveOccurred())
				Expect(serviceURL).To(BeNil())
			})
		})
	})

	Context("Environment variable resolver", func() {
		When(`special "HOST" and "PORT" environment variables are set`, func() {
			It("should form url without error", func() {
				var (
					serviceName    = "srv-name"
					portName       = "prt_name"
					hostEnvVarName = "SRV_NAME_SERVICE_HOST"
					portEnvVarName = "SRV_NAME_SERVICE_PORT_PRT_NAME"
				)

				_ = os.Setenv(hostEnvVarName, "1.2.3.4")
				_ = os.Setenv(portEnvVarName, "12345")

				resolver := discovery.NewEnvResolver()
				serviceURL, err := resolver.Resolve(portName, serviceName)

				Expect(err).ShouldNot(HaveOccurred())

				serviceURL.Scheme = "http"
				Expect(serviceURL.String()).To(Equal("http://1.2.3.4:12345"))
			})
		})
	})

	Context("Fallback resolver", func() {
		When("one of the combined resolvers fails", func() {
			It("should try the next one", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				failedResolver := mock_discovery.NewMockResolver(ctrl)
				failedResolver.
					EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("constant error"))

				successfulResolver := mock_discovery.NewMockResolver(ctrl)
				successfulResolver.
					EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Return(&url.URL{
						Scheme: "http",
						Host:   "example.com:12345",
					}, nil)

				shouldNeverCallResolver := mock_discovery.NewMockResolver(ctrl)

				fallbackResolver := discovery.NewFallbackResolver(
					failedResolver,
					successfulResolver,
					shouldNeverCallResolver,
				)
				serviceURL, err := fallbackResolver.Resolve("any", "thing")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(serviceURL.String()).To(Equal("http://example.com:12345"))
			})
		})
	})

	Context("Scheme decorator", func() {
		When("url scheme needs to be replaced", func() {
			It("should replace scheme by port name mapping", func() {
				ctrl := gomock.NewController(GinkgoT())
				defer ctrl.Finish()

				topResolver := mock_discovery.NewMockResolver(ctrl)
				topResolver.
					EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_, _ string) (*url.URL, error) {
						return &url.URL{Host: "example.com:12345", Scheme: "tcp"}, nil
					}).
					AnyTimes()

				portSchemeMapping := map[string]string{
					"api":         "http",
					"private-api": "https",
				}

				decorator := discovery.NewSchemeDecorator(topResolver, portSchemeMapping)

				apiUrl, _ := decorator.Resolve("api", "service")
				Expect(apiUrl.String()).To(Equal("http://example.com:12345"))

				privateApiUrl, _ := decorator.Resolve("private-api", "service")
				Expect(privateApiUrl.String()).To(Equal("https://example.com:12345"))

				notMappedUrl, _ := decorator.Resolve("not-mapped", "service")
				Expect(notMappedUrl.String()).To(Equal("tcp://example.com:12345"))
			})
		})
	})
})
