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

package machineinfo

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-kni/debug-tools/internal/cli/knit"
	"github.com/openshift-kni/debug-tools/pkg/machineinformer"
)

type machineInfoOptions struct {
	handle machineinformer.Handle
}

func NewMachineInfoCommand(knitOpts *knit.KnitOptions) *cobra.Command {
	opts := &machineInfoOptions{
		handle: machineinformer.Handle{
			Out: os.Stdout,
		},
	}
	mInfo := &cobra.Command{
		Use:   "machineinfo",
		Short: "show cadvisor's machine info",
		RunE: func(cmd *cobra.Command, args []string) error {
			// we need to do this AFTER we parsed the flags
			opts.handle.RootDirectory = knitOpts.SysFSRoot
			return showMachineInfo(cmd, opts, args)
		},
		Args: cobra.MaximumNArgs(1),
	}
	mInfo.Flags().BoolVarP(&opts.handle.RawOutput, "raw-output", "X", false, "include machine-identifiable data")
	mInfo.Flags().BoolVar(&opts.handle.CleanTimestamp, "clean-timestamp", false, "clean the timestamp (for testing purposes)")
	mInfo.Flags().BoolVar(&opts.handle.CleanProcfsInfo, "clean-procfs-info", false, "clean the information coming for /proc (for testing purposes)")
	return mInfo
}

func showMachineInfo(cmd *cobra.Command, opts *machineInfoOptions, args []string) error {
	opts.handle.Run()
	return nil
}
