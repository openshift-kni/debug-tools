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
 * Copyright 2020 Red Hat, Inc.
 */

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

type knitOptions struct {
	cpuList    string
	cpus       cpuset.CPUSet
	procFSRoot string
	sysFSRoot  string

	debug bool
	log   *log.Logger
}

var knitOpts knitOptions

// NewRootCommand returns entrypoint command to interact with all other commands
func NewRootCommand() *cobra.Command {

	root := &cobra.Command{
		Use:   "knit",
		Short: "knit allows to check system settings for low-latency workload",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			knitOpts.cpus, err = cpuset.Parse(knitOpts.cpuList)
			if err != nil {
				return fmt.Errorf("error parsing %q: %v", knitOpts.cpuList, err)
			}

			if knitOpts.debug {
				knitOpts.log = log.New(os.Stderr, "knit", log.LstdFlags)
			} else {
				knitOpts.log = log.New(ioutil.Discard, "", 0)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStderr(), cmd.UsageString())
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// see https://man7.org/linux/man-pages/man7/cpuset.7.html#FORMATS for more details
	root.PersistentFlags().StringVarP(&knitOpts.cpuList, "cpulist", "C", "0-16383", "isolated cpu set to check (see man (7) cpuset - List format")
	root.PersistentFlags().StringVarP(&knitOpts.procFSRoot, "procfs", "P", "/proc", "procfs root")
	root.PersistentFlags().StringVarP(&knitOpts.sysFSRoot, "sysfs", "S", "/sys", "sysfs root")
	root.PersistentFlags().BoolVarP(&knitOpts.debug, "debug", "D", false, "enable debug log")

	root.AddCommand(
		newCPUAffinityCommand(),
		newIRQAffinityCommand(),
	)

	return root

}