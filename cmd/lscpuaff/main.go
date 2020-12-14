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
	"os"
	"path/filepath"
	"sort"
	"strconv"

	flag "github.com/spf13/pflag"

	"github.com/openshift-kni/debug-tools/pkg/procs"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [cpulist]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	var procfsRoot = flag.StringP("procfs", "P", "/proc", "procfs mount point to use.")
	var pidIdent = flag.StringP("pid", "p", "", "monitor only threads belonging to this pid (default is all).")
	flag.Parse()

	isolCpuList := "0-65535" // "everything"
	args := flag.Args()
	if len(args) == 1 {
		isolCpuList = args[0]
	}

	isolCpus, err := cpuset.Parse(isolCpuList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing %q: %v", isolCpuList, err)
		os.Exit(1)
	}

	if *pidIdent != "" {
		pid, err := strconv.Atoi(*pidIdent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing %q: %v", *pidIdent, err)
			os.Exit(2)
		}
		procInfo, err := procs.FromPID(*procfsRoot, pid)
		for tid, tidInfo := range procInfo.TIDs {
			threadCpus := cpuset.NewCPUSet(tidInfo.Affinity...)

			cpus := threadCpus.Intersection(isolCpus)
			if cpus.Size() != 0 {
				fmt.Printf("PID %6d TID %6d can run on %v\n", pid, tid, cpus.ToSlice())
			}
		}
	}

	procInfos, err := procs.All(*procfsRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting process infos from %q: %v", *procfsRoot, err)
		os.Exit(1)
	}

	pids := make([]int, len(procInfos))
	for pid := range procInfos {
		pids = append(pids, pid)
	}
	sort.Ints(pids)
	for _, pid := range pids {
		procInfo := procInfos[pid]

		tids := make([]int, len(procInfo.TIDs))
		for tid := range procInfo.TIDs {
			tids = append(tids, tid)
		}
		sort.Ints(tids)

		for _, tid := range tids {
			tidInfo := procInfo.TIDs[tid]
			threadCpus := cpuset.NewCPUSet(tidInfo.Affinity...)

			cpus := threadCpus.Intersection(isolCpus)
			if cpus.Size() != 0 {
				fmt.Printf("PID %6d (%-32s) TID %6d (%-16s) can run on %v\n", pid, procInfo.Name, tid, tidInfo.Name, cpus.ToSlice())
			}
		}
	}
}
