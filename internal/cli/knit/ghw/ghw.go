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
 * Copyright 2021 Red Hat, Inc.
*/

package ghw

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/option"
	"github.com/jaypipes/ghw/pkg/pci"
	"github.com/jaypipes/ghw/pkg/topology"

	"github.com/openshift-kni/debug-tools/internal/cli/knit"
)

func NewLstopoCommand(knitOpts *knit.KnitOptions) *cobra.Command {
	topo := &cobra.Command{
		Use:   "lstopo",
		Short: "show the system topology",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ghwOptionsFromKnit(knitOpts)
			info, err := topology.New(opts...)
			return processInfo(knitOpts, info, err)
		},
		Args: cobra.NoArgs,
	}
	return topo
}

func NewLscpuCommand(knitOpts *knit.KnitOptions) *cobra.Command {
	topo := &cobra.Command{
		Use:   "lscpu",
		Short: "show the system CPU details",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ghwOptionsFromKnit(knitOpts)
			info, err := cpu.New(opts...)
			return processInfo(knitOpts, info, err)
		},
		Args: cobra.NoArgs,
	}
	return topo
}

func NewLspciCommand(knitOpts *knit.KnitOptions) *cobra.Command {
	topo := &cobra.Command{
		Use:   "lspci",
		Short: "show the system PCI details",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ghwOptionsFromKnit(knitOpts)
			info, err := pci.New(opts...)
			return processInfo(knitOpts, info, err)
		},
		Args: cobra.NoArgs,
	}
	return topo
}

type encodable interface {
	JSONString(bool) string
	YAMLString() string
}

func processInfo(knitOpts *knit.KnitOptions, info encodable, err error) error {
	if err == nil {
		if knitOpts.JsonOutput {
			fmt.Printf("%s\n", info.JSONString(true))
		} else {
			fmt.Printf("%s\n", info.YAMLString())
		}
	}
	return err
}

func ghwOptionsFromKnit(knitOpts *knit.KnitOptions) []*option.Option {
	return []*option.Option{
		option.WithPathOverrides(option.PathOverrides{
			"/proc": knitOpts.ProcFSRoot,
			"/sys":  knitOpts.SysFSRoot,
		}),
	}
}
