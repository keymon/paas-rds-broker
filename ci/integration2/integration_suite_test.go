package integration2_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/phayes/freeport"

	rdsbroker "github.com/alphagov/paas-rds-broker/aaa"

	. "github.com/alphagov/paas-rds-broker/ci/helpers"
)

var (
	rdsBrokerPath    string
	rdsBrokerPort    int
	rdsBrokerUrl     string
	rdsBrokerSession *gexec.Session

	brokerAPIClient *BrokerAPIClient

	rdsClient *RDSClient
	config    *rdsbroker.Config
)
var _ = SynchronizedBeforeSuite(func() []byte {
	var err error

	// Compile the broker
	gp, err := gexec.Build("github.com/alphagov/paas-rds-broker")
	Expect(err).ShouldNot(HaveOccurred())

	gpDir := filepath.Dir(gp)
	cp := filepath.Join(gpDir, "paas-rds-broker")

	// start the broker in a random port
	rdsBrokerPort = freeport.GetPort()
	command := exec.Command(cp, fmt.Sprintf("-port=%d", rdsBrokerPort), "-config=./config.json")
	rdsBrokerSession, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())

	// Wait for it to be listening
	Eventually(rdsBrokerSession, 5*time.Second).Should(gbytes.Say(fmt.Sprintf("RDS Service Broker started on port %d", rdsBrokerPort)))

	return nil

}, func(data []byte) {
	var err error
	rdsBrokerUrl = fmt.Sprintf("http://localhost:%d", rdsBrokerPort)
	config, err = rdsbroker.LoadConfig("./config.json")

	brokerAPIClient = NewBrokerAPIClient(rdsBrokerUrl, config.Username, config.Password)

	Expect(err).ToNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	rdsBrokerSession.Kill()
})

func TestAcceptance(t *testing.T) {
	var err error

	RegisterFailHandler(Fail)

	// FIXME: Remove hardcoded region and prefix
	rdsClient, err = NewRDSClient("eu-west-1", "rdsbroker-test")
	Expect(err).ShouldNot(HaveOccurred())

	if ok, err := rdsClient.Ping(); ok {
		RunSpecs(t, "RDS Broker Integration Suite")
	} else {
		errorName := strings.SplitN(err.Error(), ":", 2)[0]
		if errorName == "NoCredentialProviders" {
			fmt.Fprintf(os.Stderr, "WARNING: Skipping RDS Broker integration, as no credentials were provided:\n  %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: Cannot run RDS Broker integration, no credentials were provided:\n  %v", err)
			os.Exit(1)
		}
	}
}
