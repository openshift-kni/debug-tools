/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2022 Red Hat, Inc.
 */

package kubeletpods

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	corev1 "k8s.io/api/core/v1"
)

const (
	saToken = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

func KubeletURL(node string, port int) string {
	return fmt.Sprintf("https://%s:%d/pods", node, port)
}

func DefaultKubeletURL() string {
	return KubeletURL("localhost", 10250)
}

func List(podUrl string) ([]corev1.Pod, error) {
	resp, err := httpGet(podUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get the HTTP response: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %w", err)
	}
	podList := corev1.PodList{}
	err = json.Unmarshal(body, &podList)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the response body: %w", err)
	}

	return podList.Items, nil
}

func httpGet(url string) (*http.Response, error) {
	objToken, err := ioutil.ReadFile(saToken)
	if err != nil {
		return nil, fmt.Errorf("failed to read from %q: %w", saToken, err)
	}
	token := string(objToken)

	var bearer = "Bearer " + token
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", bearer)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response from %q: %w", url, err)
	}
	return resp, err
}
