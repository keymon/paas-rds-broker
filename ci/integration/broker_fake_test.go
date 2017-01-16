package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	main "github.com/alphagov/paas-rds-broker"
	rdsfake "github.com/alphagov/paas-rds-broker/awsrds/fakes"
	"github.com/alphagov/paas-rds-broker/rdsbroker"
	sqlfake "github.com/alphagov/paas-rds-broker/sqlengine/fakes"
)

type ByServiceID []brokerapi.Service

func (a ByServiceID) Len() int           { return len(a) }
func (a ByServiceID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByServiceID) Less(i, j int) bool { return a[i].ID < a[j].ID }

var _ = Describe("RDS Broker", func() {

	var (
		allowUserProvisionParameters bool
		allowUserUpdateParameters    bool
		allowUserBindParameters      bool
		serviceBindable              bool
		planUpdateable               bool
		skipFinalSnapshot            bool
		dbPrefix                     string
		brokerName                   string
		config                       *main.Config
		dbInstance                   *rdsfake.FakeDBInstance

		sqlProvider *sqlfake.FakeProvider
		sqlEngine   *sqlfake.FakeSQLEngine

		testSink *lagertest.TestSink
		logger   lager.Logger

		rdsBroker *rdsbroker.RDSBroker

		rdsBrokerServer http.Handler
	)
	const ()

	BeforeEach(func() {
		allowUserProvisionParameters = true
		allowUserUpdateParameters = true
		allowUserBindParameters = true
		serviceBindable = true
		planUpdateable = true
		skipFinalSnapshot = true
		dbPrefix = "cf"
		brokerName = "mybroker"

		dbInstance = &rdsfake.FakeDBInstance{}
		sqlProvider = &sqlfake.FakeProvider{}
		sqlEngine = &sqlfake.FakeSQLEngine{}
		sqlProvider.GetSQLEngineSQLEngine = sqlEngine

	})

	JustBeforeEach(func() {
		var err error

		config, err = main.LoadConfig("./config.json")
		Expect(err).ToNot(HaveOccurred())

		logger = lager.NewLogger("rdsbroker_test")
		testSink = lagertest.NewTestSink()
		logger.RegisterSink(testSink)

		rdsBroker = rdsbroker.New(*config.RDSConfig, dbInstance, sqlProvider, logger)
		rdsBrokerServer = main.BuildHTTPHandler(rdsBroker, logger, config)
	})

	var _ = Describe("Services", func() {
		It("returns the proper CatalogResponse", func() {
			var err error

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest("GET", "http://example.com/v2/catalog", nil)
			req.SetBasicAuth(config.Username, config.Password)

			rdsBrokerServer.ServeHTTP(recorder, req)

			catalog := brokerapi.CatalogResponse{}
			err = json.Unmarshal(recorder.Body.Bytes(), &catalog)
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

	var _ = Describe("Provision", func() {
		var (
			provisionDetailsJson []byte
			serviceID            string
			acceptsIncomplete    bool
		)

		BeforeEach(func() {
			provisionDetailsJson = []byte(`
				{
					"service_id": "Service-1",
					"plan_id": "Plan-1",
					"organization_guid": "organization-id",
					"space_guid": "space-id",
					"parameters": {}
				}
			`)
			serviceID = "Service-1"
			acceptsIncomplete = true
		})

		var doProvisionRequest = func() *httptest.ResponseRecorder {
			recorder := httptest.NewRecorder()

			path := "/v2/service_instances/" + serviceID

			if acceptsIncomplete {
				path = path + "?accepts_incomplete=true"
			}

			req := httptest.NewRequest("PUT", path, bytes.NewBuffer(provisionDetailsJson))
			req.SetBasicAuth(config.Username, config.Password)

			rdsBrokerServer.ServeHTTP(recorder, req)

			return recorder
		}

		It("returns 202 Accepted, Service instance provisioning is in progress", func() {
			recorder := doProvisionRequest()
			Expect(recorder.Code).To(Equal(202))
		})
	})

})
