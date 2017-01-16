package integration_test

import (
	"path/filepath"
	"testing"

	. "github.com/alphagov/paas-rds-broker/ci/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var rdsBrokerPath string
var rdsClient RDSClient

var _ = SynchronizedBeforeSuite(func() []byte {
	gp, err := gexec.Build("github.com/alphagov/paas-rds-broker")
	Expect(err).ShouldNot(HaveOccurred())

	gpDir := filepath.Dir(gp)
	cp := filepath.Join(gpDir, "paas-rds-broker")

	return []byte(cp)
}, func(data []byte) {
	rdsBrokerPath = string(data)
})

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	// FIXME: Remove hardcoded region
	rdsClient = NewRdsClient("eu-west-1")
	if ok, err := rdsClient.Ping(); ok {
		RunSpecs(t, "RDS Broker Integration Suite")
	}

}
