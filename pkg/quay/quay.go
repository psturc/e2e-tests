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
	"errors"
	"fmt"
	"io"
	"net/http"
)

type QuayClient struct {
	url        string
	httpClient *http.Client
	AuthToken  string
}

func NewQuayClient(c *http.Client, authToken, url string) QuayClient {
	return QuayClient{
		httpClient: c,
		AuthToken:  authToken,
		url:        url,
	}
}

// DeleteRepository deletes specified image repository.
func (c *QuayClient) DeleteRepository(organization, imageRepository string) (bool, error) {
	url := fmt.Sprintf("%s/repository/%s/%s", c.url, organization, imageRepository)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Bearer", c.AuthToken))
	req.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 204 {
		return true, nil
	}
	if res.StatusCode == 404 {
		return false, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	data := &QuayError{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return false, err
	}
	return false, errors.New(data.ErrorMessage)
}

// DeleteRobotAccount deletes given Quay.io robot account in the organization.
func (c *QuayClient) DeleteRobotAccount(organization string, robotName string) (bool, error) {
	url := fmt.Sprintf("%s/organization/%s/robots/%s", c.url, organization, robotName)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Bearer", c.AuthToken))
	req.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 204 {
		return true, nil
	}
	if res.StatusCode == 404 {
		return false, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	data := &QuayError{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return false, err
	}
	return false, errors.New(data.ErrorMessage)
}

// Returns all repositories of the DEFAULT_QUAY_ORG organization
func (c *QuayClient) GetAllRepositories(organization string) ([]Repository, error) {
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
	json.Unmarshal(body, &response)
	return response.Repositories, nil
}

// Returns all robot accounts of the DEFAULT_QUAY_ORG organization
func (c *QuayClient) GetAllRobotAccounts(organization string) ([]RobotAccount, error) {
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
		Robots []RobotAccount
	}
	var response Response
	json.Unmarshal(body, &response)
	return response.Robots, nil
}
