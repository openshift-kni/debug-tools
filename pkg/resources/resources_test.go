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

package resources

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/utils/cpuset"

	"github.com/openshift-kni/debug-tools/pkg/cgroups"
	"github.com/openshift-kni/debug-tools/pkg/environ"
)

func TestDiscover(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		content     string
		expectedRes Resources
		expectedErr bool
	}{
		{
			name:        "non-existent path",
			path:        "/this/path/does/not/exist",
			expectedErr: true,
		},
		{
			name:    "simple happy path",
			content: "0-9",
			expectedRes: Resources{
				CPUs: cpuset.New(0, 1, 2, 3, 4, 5, 6, 7, 8, 9),
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var env environ.Environ
			if len(tt.path) > 0 {
				env = environ.Environ{
					Root: environ.FS{
						Sys: tt.path,
					},
					Log: environ.DefaultLog(),
				}
			} else if len(tt.content) > 0 {
				tmpDir := t.TempDir()
				env = environ.Environ{
					Root: environ.FS{
						Sys: tmpDir,
					},
					Log: environ.DefaultLog(),
				}
				tmpPath := cgroups.CpusetPath(&env)
				err := os.MkdirAll(filepath.Dir(tmpPath), os.ModePerm)
				if err != nil {
					t.Fatalf("cannot prepare the fake data path at %v: %v", tmpPath, err)
				}
				err = os.WriteFile(tmpPath, []byte(tt.content), 0o644)
				if err != nil {
					t.Fatalf("cannot prepare the fake data file at %v: %v", tmpPath, err)
				}
			} else {
				t.Fatalf("neither path or content given; wrong test")
			}

			got, err := Discover(&env)
			if tt.expectedErr && err == nil {
				t.Fatalf("expected error, got success")
			}
			if !tt.expectedErr && err != nil {
				t.Fatalf("expected success, got err=%v", err)
			}
			if tt.expectedRes.CPUs.Size() > 0 && !got.CPUs.Equals(tt.expectedRes.CPUs) {
				t.Fatalf("expected Resources %v got %v", tt.expectedRes, got)
			}
		})
	}
}
