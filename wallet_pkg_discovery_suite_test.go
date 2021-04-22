package discovery_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWalletPkgDiscovery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WalletPkgDiscovery Suite")
}
