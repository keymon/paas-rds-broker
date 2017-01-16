package integration2_test

import (
	"encoding/json"
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

	main "github.com/alphagov/paas-rds-broker"

	. "github.com/alphagov/paas-rds-broker/ci/helpers"
)

type suiteDataStruct struct {
	rdsBrokerPath string
	port          int
	session       *gexec.Session
}

var suiteData suiteDataStruct
var rdsBrokerUrl string
var rdsClient *RDSClient
var config *main.Config

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error

	// Compile the broker
	gp, err := gexec.Build("github.com/alphagov/paas-rds-broker")
	Expect(err).ShouldNot(HaveOccurred())

	gpDir := filepath.Dir(gp)
	cp := filepath.Join(gpDir, "paas-rds-broker")

	// start the broker in a random port
	port := freeport.GetPort()
	command := exec.Command(cp, fmt.Sprintf("-port=%d", port), "-config=./config.json")
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())

	// Wait for it to be listening
	Eventually(session, 5*time.Second).Should(gbytes.Say(fmt.Sprintf("RDS Service Broker started on port %d", port)))

	// Pass the data to the workers
	data, err := json.Marshal(suiteDataStruct{
		rdsBrokerPath: cp,
		port:          port,
		session:       session,
	})
	Expect(err).ShouldNot(HaveOccurred())

	return data

}, func(data []byte) {
	var err error

	fmt.Fprintf(os.Stderr, "data: %v", string(data))

	err = json.Unmarshal(data, &suiteData)
	Expect(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(os.Stderr, "port: %v", suiteData)

	rdsBrokerUrl = fmt.Sprintf("http://localhost:%d", suiteData.port)

	config, err = main.LoadConfig("./config.json")
	Expect(err).ToNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	suiteData.session.Kill()
})

func TestAcceptance(t *testing.T) {
	var err error

	RegisterFailHandler(Fail)

	// FIXME: Remove hardcoded region
	rdsClient, err = NewRDSClient("eu-west-1")
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
