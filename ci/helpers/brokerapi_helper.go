package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/frodenas/brokerapi"
)

type ByServiceID []brokerapi.Service

func (a ByServiceID) Len() int           { return len(a) }
func (a ByServiceID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByServiceID) Less(i, j int) bool { return a[i].ID < a[j].ID }

func BodyBytes(resp *http.Response) ([]byte, error) {
	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

type BrokerAPIClient struct {
	Url               string
	Username          string
	Password          string
	AcceptsIncomplete bool
}

func NewBrokerAPIClient(Url string, Username string, Password string) *BrokerAPIClient {
	return &BrokerAPIClient{
		Url:      Url,
		Username: Username,
		Password: Password,
	}
}

func (b *BrokerAPIClient) doRequest(action string, path string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(action, b.Url+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(b.Username, b.Password)

	return client.Do(req)
}

func (b *BrokerAPIClient) GetCatalog() (brokerapi.CatalogResponse, error) {

	catalog := brokerapi.CatalogResponse{}

	resp, err := b.doRequest("GET", "/v2/catalog", nil)
	if err != nil {
		return catalog, err
	}
	if resp.StatusCode != 200 {
		return catalog, fmt.Errorf("Invalid catalog response %v", resp)
	}

	body, err := BodyBytes(resp)
	if err != nil {
		return catalog, err
	}

	err = json.Unmarshal(body, &catalog)
	if err != nil {
		return catalog, err
	}

	return catalog, nil
}

func (b *BrokerAPIClient) DoProvisionRequest(serviceID string, planID string) (*http.Response, error) {
	path := "/v2/service_instances/" + serviceID

	if b.AcceptsIncomplete {
		path = path + "?accepts_incomplete=true"
	}

	provisionDetailsJson := []byte(fmt.Sprintf(`
		{
			"service_id": "%s",
			"plan_id": "%s",
			"organization_guid": "test-organization-id",
			"space_guid": "space-id",
			"parameters": {}
		}
	`, serviceID, planID))

	return b.doRequest("PUT", path, bytes.NewBuffer(provisionDetailsJson))
}

func (b *BrokerAPIClient) DoDeprovisionRequest(serviceID string, planID string) (*http.Response, error) {
	path := "/v2/service_instances/" + serviceID

	if b.AcceptsIncomplete {
		path = path + "?accepts_incomplete=true"
	}

	provisionDetailsJson := []byte(fmt.Sprintf(`
		{
			"service_id": "%s",
			"plan_id": "%s",
			"organization_guid": "test-organization-id",
			"space_guid": "space-id",
			"parameters": {}
		}
	`, serviceID, planID))

	return b.doRequest("DELETE", path, bytes.NewBuffer(provisionDetailsJson))
}

//func pollForRDSCreationCompletion(dbInstanceName string) {
//fmt.Fprint(GinkgoWriter, "Polling for RDS creation to complete")
//Eventually(func() *Buffer {
//fmt.Fprint(GinkgoWriter, ".")
//command := quietCf("cf", "service", dbInstanceName).Wait(DEFAULT_TIMEOUT)
//Expect(command).To(Exit(0))
//return command.Out
//}, DB_CREATE_TIMEOUT, 15*time.Second).Should(Say("create succeeded"))
//fmt.Fprint(GinkgoWriter, "done\n")
//}
