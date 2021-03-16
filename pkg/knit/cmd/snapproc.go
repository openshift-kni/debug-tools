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
	"os"

	"github.com/jaypipes/ghw/pkg/snapshot"
	"github.com/spf13/cobra"

	"github.com/openshift-kni/debug-tools/pkg/procs"
)

type snapProcOptions struct {
	preserve     bool
	cloneDirPath string
	output       string
}

func NewSnapProcCommand(knitOpts *KnitOptions) *cobra.Command {
	opts := &snapProcOptions{}
	snapProc := &cobra.Command{
		Use:   "snapproc",
		Short: "create snapshot of running processes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return makeProcSnapshot(cmd, knitOpts, opts, args)
		},
		Args: cobra.NoArgs,
	}
	snapProc.Flags().BoolVar(&opts.preserve, "preserve", false, "do not remove the scratch directory clone once done.")
	snapProc.Flags().StringVarP(&opts.cloneDirPath, "clone-path", "c", "", "proc clone path. If not given, generate random name.")
	snapProc.Flags().StringVarP(&opts.output, "output", "o", "snapproc.tgz", "proc snapshot. Use \"-\" to send on stdout.")
	return snapProc
}

func makeProcSnapshot(cmd *cobra.Command, knitOpts *KnitOptions, opts *snapProcOptions, args []string) error {
	var scratchDir string
	var err error
	if opts.cloneDirPath != "" {
		// if user supplies the path, then the user owns the path. So we intentionally don't remove it once done.
		scratchDir = opts.cloneDirPath
	} else {
		scratchDir, err = ioutil.TempDir("", "kni-snapproc")
		if err != nil {
			return err
		}
		if opts.preserve {
			defer fmt.Fprintf(os.Stderr, "preserved scratch directory %q\n", scratchDir)
		} else {
			defer os.RemoveAll(scratchDir)
		}
	}

	if err = procs.CloneInto(knitOpts.Log, knitOpts.ProcFSRoot, scratchDir); err != nil {
		return err
	}

	return snapshot.PackFrom(opts.output, scratchDir)
}
