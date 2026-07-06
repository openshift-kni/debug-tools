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

package knit

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

type goProcsOptions struct{}

func NewGOProcsCommand(knitOpts *KnitOptions) *cobra.Command {
	opts := &goProcsOptions{}
	goProcs := &cobra.Command{
		Use:   "goprocs",
		Short: "Show golang concurrency settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showGOProcs(cmd, knitOpts, opts, args)
		},
		Args: cobra.NoArgs,
	}
	return goProcs
}

func showGOProcs(cmd *cobra.Command, knitOpts *KnitOptions, opts *goProcsOptions, args []string) error {
	fmt.Printf("%d:%d\n", runtime.GOMAXPROCS(0), runtime.NumCPU())
	return nil
}
