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
 * Copyright 2020-2021 Red Hat, Inc.
 */

package main

import (
	"fmt"
	"os"

	"github.com/openshift-kni/debug-tools/pkg/knit/cmd"
	"github.com/openshift-kni/debug-tools/pkg/knit/cmd/ghw"
	"github.com/openshift-kni/debug-tools/pkg/knit/cmd/k8s"
)

func main() {
	root := cmd.NewRootCommand(
		k8s.NewPodResourcesCommand,
		ghw.NewLscpuCommand,
		ghw.NewLspciCommand,
		ghw.NewLstopoCommand,
	)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
