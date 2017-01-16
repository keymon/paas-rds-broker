package integration2_test

import (
	"sort"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	uuid "github.com/satori/go.uuid"

	. "github.com/alphagov/paas-rds-broker/ci/helpers"
)

var _ = Describe("RDS Broker Daemon", func() {
	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	FIt("should check the instance credentials", func() {
		Eventually(rdsBrokerSession, 30*time.Second).Should(gbytes.Say("credentials check has ended"))
	})

	var _ = FDescribe("Services", func() {
		It("returns the proper CatalogResponse", func() {
			var err error

			catalog, err := brokerAPIClient.GetCatalog()
			Expect(err).ToNot(HaveOccurred())

			sort.Sort(ByServiceID(catalog.Services))

			Expect(catalog.Services).To(HaveLen(3))
			service1 := catalog.Services[0]
			service2 := catalog.Services[1]
			service3 := catalog.Services[2]
			Expect(service1.ID).To(Equal("Service-1"))
			Expect(service2.ID).To(Equal("Service-2"))
			Expect(service3.ID).To(Equal("Service-3"))

			Expect(service1.ID).To(Equal("Service-1"))
			Expect(service1.Name).To(Equal("Service 1"))
			Expect(service1.Description).To(Equal("This is the Service 1"))
			Expect(service1.Bindable).To(BeTrue())
			Expect(service1.PlanUpdateable).To(BeTrue())
			Expect(service1.Plans).To(HaveLen(1))
			Expect(service1.Plans[0].ID).To(Equal("Plan-1"))
			Expect(service1.Plans[0].Name).To(Equal("Plan 1"))
			Expect(service1.Plans[0].Description).To(Equal("This is the Plan 1"))
		})
	})

	var _ = FDescribe("Instance Provision/Update/Deprovision", func() {
		var (
			serviceID string
			planID    string
		)

		BeforeEach(func() {
			serviceID = uuid.NewV4().String()
			planID = "Plan-1"

			brokerAPIClient.AcceptsIncomplete = true

			resp, err := brokerAPIClient.DoProvisionRequest(serviceID, planID)
			Expect(resp.StatusCode).To(Equal(202))

			Expect(err).ToNot(HaveOccurred())
			// TODO poll
		})

		AfterEach(func() {
			brokerAPIClient.AcceptsIncomplete = true
			resp, err := brokerAPIClient.DoDeprovisionRequest(serviceID, planID)
			Expect(resp.StatusCode).To(Equal(202))
			Expect(err).ToNot(HaveOccurred())
			// pollForRDSDeletionCompletion(dbInstanceName)
		})

		It("aaa", func() {
			Expect(1).To(Equal(2))
		})

	})
})
