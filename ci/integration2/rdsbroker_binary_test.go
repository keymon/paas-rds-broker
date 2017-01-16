package integration2_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/frodenas/brokerapi"
	uuid "github.com/satori/go.uuid"

	. "github.com/alphagov/paas-rds-broker/ci/helpers"
)

func BodyBytes(resp *http.Response) []byte {
	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(resp.Body)
	Expect(err).ToNot(HaveOccurred())
	return buf.Bytes()
}

var _ = Describe("RDS Broker Daemon", func() {
	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	var doRDSBrokerRequest = func(action string, path string, body io.Reader) *http.Response {
		client := &http.Client{}
		req, err := http.NewRequest(action, rdsBrokerUrl+path, body)
		Expect(err).ToNot(HaveOccurred())
		resp, err := client.Do(req)
		Expect(err).ToNot(HaveOccurred())

		return resp
	}

	It("should check the instance credentials", func() {
		Eventually(suiteData.session, 5*time.Second).Should(gbytes.Say("credentials check has ended"))
	})

	var _ = Describe("Services", func() {
		It("returns the proper CatalogResponse", func() {
			var err error

			resp := doRDSBrokerRequest("GET", "/v2/catalog", nil)
			Expect(resp.StatusCode).To(Equal(200))

			catalog := brokerapi.CatalogResponse{}
			err = json.Unmarshal(BodyBytes(resp), &catalog)
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

	var _ = Describe("Instance Provision/Update/Deprovision", func() {
		var (
			provisionDetailsJson []byte
			serviceID            string
			acceptsIncomplete    bool
		)

		var doProvisionRequest = func(serviceID string) *http.Response {
			path := "/v2/service_instances/" + serviceID

			if acceptsIncomplete {
				path = path + "?accepts_incomplete=true"
			}

			resp := doRDSBrokerRequest("PUT", path, bytes.NewBuffer(provisionDetailsJson))
			Expect(resp.StatusCode).To(Equal(202))

			return resp
		}

		var doDeprovisionRequest = func(serviceID string) *http.Response {
			path := "/v2/service_instances/" + serviceID

			if acceptsIncomplete {
				path = path + "?accepts_incomplete=true"
			}

			resp := doRDSBrokerRequest("DELETE", path, bytes.NewBuffer(provisionDetailsJson))
			Expect(resp.StatusCode).To(Equal(202))

			return resp
		}

		BeforeEach(func() {
			serviceID = uuid.NewV4().String()
			provisionDetailsJson = []byte(fmt.Sprintf(`
				{
					"service_id": "%s",
					"plan_id": "Plan-1",
					"organization_guid": "test-organization-id",
					"space_guid": "space-id",
					"parameters": {}
				}
			`, serviceID))
			acceptsIncomplete = true

			doProvisionRequest(serviceID)
			// TODO poll
		})

		AfterEach(func() {
			if false {
				doDeprovisionRequest(serviceID)
			}
			// pollForRDSDeletionCompletion(dbInstanceName)
		})

		It("aaa", func() {
			Expect(1).To(Equal(2))
		})

	})
})
