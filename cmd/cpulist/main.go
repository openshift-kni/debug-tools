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

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/openshift-kni/debug-tools/pkg/procs"
	cpuset "k8s.io/utils/cpuset"
)

func main() {
	var procfsRoot = flag.StringP("procfs", "P", "/proc", "procfs root")
	var cpuList = flag.StringP("cpu-list", "c", "", "cpulist to split")
	var srcFile = flag.StringP("from-file", "f", "", "read the cpulist to split from the given file")
	flag.Parse()

	var cpus cpuset.CPUSet

	if *srcFile != "" {
		var err error
		var data []byte
		if *srcFile == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(*srcFile)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading cpulist from %q: %v\n", *srcFile, err)
			os.Exit(2)
		}
		cpus = parseCPUsOrDie(strings.TrimSpace(string(data)))
	} else if *cpuList != "" {
		cpus = parseCPUsOrDie(*cpuList)
	} else {
		cpus = allowedCPUsOrDie(*procfsRoot)
	}
	printCPUList(cpus)
}

func parseCPUsOrDie(cpuList string) cpuset.CPUSet {
	cpus, err := cpuset.Parse(cpuList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing %q: %v\n", cpuList, err)
		os.Exit(2)
	}
	return cpus
}

func allowedCPUsOrDie(procfsRoot string) cpuset.CPUSet {
	nullLog := log.New(io.Discard, "", 0)
	ph := procs.New(nullLog, procfsRoot)
	info, err := ph.FromPID(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading process status for pid self: %v\n", err)
		os.Exit(4)
	}
	// consolidate all the cpus:
	var cpuIDs []int
	for _, tidInfo := range info.TIDs {
		for _, cpuId := range tidInfo.Affinity {
			cpuIDs = append(cpuIDs, cpuId)
		}
	}
	return cpuset.New(cpuIDs...)
}

func printCPUList(cpus cpuset.CPUSet) {
	for _, cpu := range cpus.List() {
		fmt.Printf("%v\n", cpu)
	}
}
