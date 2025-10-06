/*
 * Copyright 2025 Red Hat, Inc.
 *
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
 */

package k8sclient

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Create() (kubernetes.Interface, error) {
	var err error
	var kubeconfig *rest.Config

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath != "" {
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		kubeconfig, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("unable to load kubeconfig (path=%s)", kubeconfigPath)
	}

	return kubernetes.NewForConfig(kubeconfig)
}
