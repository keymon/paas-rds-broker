package integration_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/frodenas/brokerapi"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	rdsfake "github.com/alphagov/paas-rds-broker/awsrds/fakes"
	"github.com/alphagov/paas-rds-broker/rdsbroker"
	sqlfake "github.com/alphagov/paas-rds-broker/sqlengine/fakes"
)

func buildHTTPHandler(serviceBroker *rdsbroker.RDSBroker, logger lager.Logger, user string, pass string) http.Handler {
	credentials := brokerapi.BrokerCredentials{
		Username: user,
		Password: pass,
	}

	brokerAPI := brokerapi.New(serviceBroker, logger, credentials)
	mux := http.NewServeMux()
	mux.Handle("/", brokerAPI)
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

var _ = Describe("RDS Broker", func() {
	var (
		rdsProperties1 rdsbroker.RDSProperties
		rdsProperties2 rdsbroker.RDSProperties
		rdsProperties3 rdsbroker.RDSProperties
		plan1          rdsbroker.ServicePlan
		plan2          rdsbroker.ServicePlan
		plan3          rdsbroker.ServicePlan
		service1       rdsbroker.Service
		service2       rdsbroker.Service
		service3       rdsbroker.Service
		catalog        rdsbroker.Catalog

		config rdsbroker.Config

		dbInstance *rdsfake.FakeDBInstance

		sqlProvider *sqlfake.FakeProvider
		sqlEngine   *sqlfake.FakeSQLEngine

		testSink *lagertest.TestSink
		logger   lager.Logger

		rdsBroker *rdsbroker.RDSBroker

		allowUserProvisionParameters bool
		allowUserUpdateParameters    bool
		allowUserBindParameters      bool
		serviceBindable              bool
		planUpdateable               bool
		skipFinalSnapshot            bool
		dbPrefix                     string
		brokerName                   string
	)

	const (
		masterPasswordSeed   = "something-secret"
		instanceID           = "instance-id"
		bindingID            = "binding-id"
		dbInstanceIdentifier = "cf-instance-id"
		dbName               = "cf_instance_id"
		dbUsername           = "uvMSB820K_t3WvCX"
		masterUserPassword   = "qOeiJ6AstR_mUQJxn6jyew=="
	)

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
		rdsProperties1 = rdsbroker.RDSProperties{
			DBInstanceClass:   "db.m1.test",
			Engine:            "test-engine-1",
			EngineVersion:     "1.2.3",
			AllocatedStorage:  100,
			SkipFinalSnapshot: skipFinalSnapshot,
		}

		rdsProperties2 = rdsbroker.RDSProperties{
			DBInstanceClass:   "db.m2.test",
			Engine:            "test-engine-2",
			EngineVersion:     "4.5.6",
			AllocatedStorage:  200,
			SkipFinalSnapshot: skipFinalSnapshot,
		}

		rdsProperties3 = rdsbroker.RDSProperties{
			DBInstanceClass:   "db.m3.test",
			Engine:            "test-engine-3",
			EngineVersion:     "4.5.6",
			AllocatedStorage:  300,
			SkipFinalSnapshot: false,
		}

		plan1 = rdsbroker.ServicePlan{
			ID:            "Plan-1",
			Name:          "Plan 1",
			Description:   "This is the Plan 1",
			RDSProperties: rdsProperties1,
		}
		plan2 = rdsbroker.ServicePlan{
			ID:            "Plan-2",
			Name:          "Plan 2",
			Description:   "This is the Plan 2",
			RDSProperties: rdsProperties2,
		}
		plan3 = rdsbroker.ServicePlan{
			ID:            "Plan-3",
			Name:          "Plan 3",
			Description:   "This is the Plan 3",
			RDSProperties: rdsProperties3,
		}

		service1 = rdsbroker.Service{
			ID:             "Service-1",
			Name:           "Service 1",
			Description:    "This is the Service 1",
			Bindable:       serviceBindable,
			PlanUpdateable: planUpdateable,
			Plans:          []rdsbroker.ServicePlan{plan1},
		}
		service2 = rdsbroker.Service{
			ID:             "Service-2",
			Name:           "Service 2",
			Description:    "This is the Service 2",
			Bindable:       serviceBindable,
			PlanUpdateable: planUpdateable,
			Plans:          []rdsbroker.ServicePlan{plan2},
		}
		service3 = rdsbroker.Service{
			ID:             "Service-3",
			Name:           "Service 3",
			Description:    "This is the Service 3",
			Bindable:       serviceBindable,
			PlanUpdateable: planUpdateable,
			Plans:          []rdsbroker.ServicePlan{plan3},
		}

		catalog = rdsbroker.Catalog{
			Services: []rdsbroker.Service{service1, service2, service3},
		}

		config = rdsbroker.Config{
			Region:                       "rds-region",
			DBPrefix:                     dbPrefix,
			BrokerName:                   brokerName,
			MasterPasswordSeed:           masterPasswordSeed,
			AllowUserProvisionParameters: allowUserProvisionParameters,
			AllowUserUpdateParameters:    allowUserUpdateParameters,
			AllowUserBindParameters:      allowUserBindParameters,
			Catalog:                      catalog,
		}

		logger = lager.NewLogger("rdsbroker_test")
		testSink = lagertest.NewTestSink()
		logger.RegisterSink(testSink)

		rdsBroker = rdsbroker.New(config, dbInstance, sqlProvider, logger)
		// rdsBrokerServer := httptest.NewServer(rdsBroker)
	})

	var _ = Describe("Services", func() {
		var (
			properCatalogResponse brokerapi.CatalogResponse
		)

		BeforeEach(func() {
			properCatalogResponse = brokerapi.CatalogResponse{
				Services: []brokerapi.Service{
					brokerapi.Service{
						ID:             "Service-1",
						Name:           "Service 1",
						Description:    "This is the Service 1",
						Bindable:       serviceBindable,
						PlanUpdateable: planUpdateable,
						Plans: []brokerapi.ServicePlan{
							brokerapi.ServicePlan{
								ID:          "Plan-1",
								Name:        "Plan 1",
								Description: "This is the Plan 1",
							},
						},
					},
					brokerapi.Service{
						ID:             "Service-2",
						Name:           "Service 2",
						Description:    "This is the Service 2",
						Bindable:       serviceBindable,
						PlanUpdateable: planUpdateable,
						Plans: []brokerapi.ServicePlan{
							brokerapi.ServicePlan{
								ID:          "Plan-2",
								Name:        "Plan 2",
								Description: "This is the Plan 2",
							},
						},
					},
					brokerapi.Service{
						ID:             "Service-3",
						Name:           "Service 3",
						Description:    "This is the Service 3",
						Bindable:       serviceBindable,
						PlanUpdateable: planUpdateable,
						Plans: []brokerapi.ServicePlan{
							brokerapi.ServicePlan{
								ID:          "Plan-3",
								Name:        "Plan 3",
								Description: "This is the Plan 3",
							},
						},
					},
				},
			}
		})

		It("returns the proper CatalogResponse", func() {
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			w := httptest.NewRecorder()
			server := buildHTTPHandler(rdsBroker, logger, "aa", "bb")
			server.ServeHTTP(w, req)

			Expect(w.Body.String()).To(Equal("ejee"))
		})

	})

})
