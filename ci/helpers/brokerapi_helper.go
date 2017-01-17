package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/frodenas/brokerapi"
	"github.com/onsi/ginkgo"
)

type ByServiceID []brokerapi.Service

func (a ByServiceID) Len() int           { return len(a) }
func (a ByServiceID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByServiceID) Less(i, j int) bool { return a[i].ID < a[j].ID }

type ProvisioningResponse struct {
	DashboardURL string `json:"dashboard_url,omitempty"`
	Operation    string `json:"operation,omitempty"`
}

type LastOperationResponse struct {
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
}

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

func (b *BrokerAPIClient) doRequest(action string, path string, body io.Reader, params ...[]string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(action, b.Url+path, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(b.Username, b.Password)

	q := req.URL.Query()
	for _, p := range params {
		if len(p) > 0 {
			k := p[0]
			v := ""
			if len(p) > 1 {
				v := p[1]
			}
			p.Add(k, v)
		}
	}
	req.URL.RawQuery = q.Encode()

	fmt.Fprintf(ginkgo.GinkgoWriter, "%v\n", req)
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

	provisionDetailsJson := []byte(fmt.Sprintf(`
		{
			"service_id": "%s",
			"plan_id": "%s",
			"organization_guid": "test-organization-id",
			"space_guid": "space-id",
			"parameters": {}
		}
	`, serviceID, planID))

	return b.doRequest(
		"PUT",
		path,
		bytes.NewBuffer(provisionDetailsJson),
		[]string{"accepts_incomplete", string(b.AcceptsIncomplete)},
	)
}

func (b *BrokerAPIClient) DoDeprovisionRequest(serviceID string, planID string) (*http.Response, error) {
	path := fmt.Sprintf("/v2/service_instances/%s", serviceID)

	fmt.Fprintf(os.Stderr, path)
	os.Exit(1)

	return b.doRequest(
		"DELETE",
		path,
		bytes.NewBuffer(provisionDetailsJson),
		[]string{"service_id", serviceID},
		[]string{"plan_id", planID},
		[]string{"accepts_incomplete", string(b.AcceptsIncomplete)},
	)
}

func (b *BrokerAPIClient) ProvisionInstance(serviceID string, planID string) (responseCode int, operation string, err error) {
	resp, err := b.DoProvisionRequest(serviceID, planID)
	if err != nil {
		return resp.StatusCode, "", err
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202 {
		provisioningResponse := ProvisioningResponse{}

		body, err := BodyBytes(resp)
		if err != nil {
			return resp.StatusCode, "", err
		}

		err = json.Unmarshal(body, &provisioningResponse)
		if err != nil {
			return resp.StatusCode, "", err
		}

		return resp.StatusCode, provisioningResponse.Operation, err
	}

	return resp.StatusCode, "", nil
}

func (b *BrokerAPIClient) DeprovisionInstance(serviceID string, planID string) (responseCode int, operation string, err error) {
	resp, err := b.DoDeprovisionRequest(serviceID, planID)
	if err != nil {
		return resp.StatusCode, "", err
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202 {
		provisioningResponse := ProvisioningResponse{}

		body, err := BodyBytes(resp)
		if err != nil {
			return resp.StatusCode, "", err
		}

		err = json.Unmarshal(body, &provisioningResponse)
		if err != nil {
			return resp.StatusCode, "", err
		}

		return resp.StatusCode, provisioningResponse.Operation, err
	}

	return resp.StatusCode, "", nil
}

func (b *BrokerAPIClient) DoLastOperationRequest(serviceID string, planID string, operation string) (*http.Response, error) {
	path := fmt.Sprintf("/v2/service_instances/%s/last_operation", serviceID)

	if operation != "" {
		path = path + "&operation=" + operation
	}

	return b.doRequest(
		"DELETE",
		path,
		bytes.NewBuffer(provisionDetailsJson),
		[]string{"service_id", serviceID},
		[]string{"plan_id", planID},
	)

	return b.doRequest("GET", path, nil)
}

func (b *BrokerAPIClient) GetLastOperationState(serviceID string, planID string, operation string) (string, error) {
	resp, err := b.DoLastOperationRequest(serviceID, planID, operation)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 401 {
		return "gone", nil
	}
	lastOperationResponse := LastOperationResponse{}

	body, err := BodyBytes(resp)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &lastOperationResponse)
	if err != nil {
		return "", err
	}

	return lastOperationResponse.State, nil

}
