/*
Copyright 2023 Red Hat, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
	http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Taken from https://github.com/redhat-appstudio/image-controller/blob/e7ced110d184bdb0935a9c39bbbf9ba3d9e8b359/pkg/quay/quay.go

package quay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	ic "github.com/redhat-appstudio/image-controller/pkg/quay"
)

type Repository struct {
	*ic.Repository
	LastModified string `json:"last_modified"`
}

type E2EQuayClient struct {
	*ic.QuayClient
	httpClient *http.Client
	url        string
}

func NewE2EQuayClient(c *http.Client, authToken, url string) E2EQuayClient {
	qc := ic.NewQuayClient(c, authToken, url)
	return E2EQuayClient{
		&qc, c, url,
	}
}

// Returns all repositories of the DEFAULT_QUAY_ORG organization
func (c E2EQuayClient) GetAllRepositories(organization string) ([]Repository, error) {
	url := fmt.Sprintf("%s/repository?last_modified=true&namespace=%s", c.url, organization)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Bearer", c.AuthToken))
	req.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		fmt.Printf("error getting repositories, got status code %d", res.StatusCode)
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Repositories []Repository
	}
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response.Repositories, nil
}

// Returns all robot accounts of the DEFAULT_QUAY_ORG organization
func (c *E2EQuayClient) GetAllRobotAccounts(organization string) ([]ic.RobotAccount, error) {
	url := fmt.Sprintf("%s/organization/%s/robots", c.url, organization)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))
	req.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		fmt.Printf("error getting robot accounts, got status code %d", res.StatusCode)
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Robots []ic.RobotAccount
	}
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response.Robots, nil
}
