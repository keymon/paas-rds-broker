package integration2_test

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/phayes/freeport"

	. "github.com/alphagov/paas-rds-broker/ci/helpers"
)

var (
	rdsBrokerPath    string
	rdsBrokerPort    int
	rdsBrokerUrl     string
	rdsBrokerSession *gexec.Session

	brokerAPIClient *BrokerAPIClient

	rdsClient *RDSClient
)

func TestSuite(t *testing.T) {
	BeforeSuite(func() {
		var err error

		// Compile the broker
		cp, err := gexec.Build("github.com/alphagov/paas-rds-broker")
		Expect(err).ShouldNot(HaveOccurred())

		// start the broker in a random port
		rdsBrokerPort = freeport.GetPort()
		command := exec.Command(cp, fmt.Sprintf("-port=%d", rdsBrokerPort), "-config=./config.json")
		rdsBrokerSession, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())

		// Wait for it to be listening
		Eventually(rdsBrokerSession, 10*time.Second).Should(gbytes.Say(fmt.Sprintf("RDS Service Broker started on port %d", rdsBrokerPort)))

		rdsBrokerUrl = fmt.Sprintf("http://localhost:%d", rdsBrokerPort)

		// FIXME: Remove hardcoded region and prefix and user credentials
		brokerAPIClient = NewBrokerAPIClient(rdsBrokerUrl, "username", "password")
		rdsClient, err = NewRDSClient("eu-west-1", "rdsbroker-test")

		Expect(err).ToNot(HaveOccurred())
	})

	AfterSuite(func() {
		rdsBrokerSession.Kill()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "RDS Broker Integration Suite")
}
