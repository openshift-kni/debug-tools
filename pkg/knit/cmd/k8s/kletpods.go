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

package k8s

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"

	"github.com/openshift-kni/debug-tools/pkg/knit/cmd"
	"github.com/openshift-kni/debug-tools/pkg/kubeletpods"
)

type kletPodsOptions struct {
	guaranteed bool
}

func NewKubeletPodsCommand(knitOpts *cmd.KnitOptions) *cobra.Command {
	opts := &kletPodsOptions{}
	kletPods := &cobra.Command{
		Use:   "kletpods",
		Short: "show pods managed by this kubelet",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showKubeletPods(cmd, opts, args)
		},
		Args: cobra.NoArgs,
	}
	kletPods.Flags().BoolVarP(&opts.guaranteed, "guaranteed", "G", false, "show only the guaranteed pods")
	return kletPods
}

func showKubeletPods(cmd *cobra.Command, opts *kletPodsOptions, args []string) error {
	kletPods, err := kubeletpods.List(kubeletpods.DefaultKubeletURL())
	if err != nil {
		return err
	}
	pods := kletPods
	if opts.guaranteed {
		pods = filterGuaranteedPods(kletPods)
	}
	return json.NewEncoder(os.Stdout).Encode(pods)
}

func filterGuaranteedPods(pods []corev1.Pod) []corev1.Pod {
	var ret []corev1.Pod
	for _, pod := range pods {
		if pod.Status.QOSClass != corev1.PodQOSGuaranteed {
			continue
		}
		ret = append(ret, pod)
	}
	return ret
}
