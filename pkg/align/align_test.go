/*
 * Copyright 2024 Red Hat, Inc.
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

package align

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"

	"k8s.io/utils/cpuset"

	apiv0 "github.com/openshift-kni/debug-tools/internal/api/v0"
	"github.com/openshift-kni/debug-tools/pkg/environ"
	"github.com/openshift-kni/debug-tools/pkg/machine"
	"github.com/openshift-kni/debug-tools/pkg/resources"
)

func TestCheck(t *testing.T) {
	root, err := getRootPath()
	if err != nil {
		t.Fatalf("cannot find the root: %v", err)
	}
	machinePath := filepath.Join(root, "hack", "machine.json")
	machineData, err := os.ReadFile(machinePath)
	if err != nil {
		t.Fatalf("cannot read machine info: %v", err)
	}

	info, err := machine.FromJSON(string(machineData))
	if err != nil {
		t.Fatalf("cannot decode machine info: %v", err)
	}
	env := environ.New()

	testCases := []struct {
		name          string
		res           resources.Resources
		expectedAlloc apiv0.Allocation
		expectedErr   bool
	}{
		{
			name: "trivialest alloc",
			res: resources.Resources{
				CPUs: cpuset.New(0),
			},
			expectedAlloc: apiv0.Allocation{
				Alignment: apiv0.Alignment{
					SMT:  false,
					LLC:  true,
					NUMA: true,
				},
				Aligned: &apiv0.AlignedInfo{
					LLC: map[int]apiv0.ContainerResourcesDetails{
						0: apiv0.ContainerResourcesDetails{
							CPUs: []int{0},
						},
					},
					NUMA: map[int]apiv0.ContainerResourcesDetails{
						0: apiv0.ContainerResourcesDetails{
							CPUs: []int{0},
						},
					},
				},
				Unaligned: &apiv0.UnalignedInfo{
					SMT: apiv0.ContainerResourcesDetails{
						CPUs: []int{16},
					},
				},
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Check(env, tt.res, info)

			if err == nil && tt.expectedErr {
				t.Fatalf("got success, but expected error")
			}
			if err != nil && !tt.expectedErr {
				t.Fatalf("got error %v but expected success", err)
			}

			gotJSON := toJSON(got)
			expJSON := toJSON(tt.expectedAlloc)
			if gotJSON != expJSON {
				t.Fatalf("got=%v expected=%v", gotJSON, expJSON)
			}
		})
	}
}

func toJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<JSON marshal error: %v>", err)
	}
	return string(data)
}

func getRootPath() (string, error) {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot retrieve tests directory")
	}
	basedir := filepath.Dir(file)
	return filepath.Abs(filepath.Join(basedir, "..", ".."))
}
