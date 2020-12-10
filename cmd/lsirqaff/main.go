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

	flag "github.com/spf13/pflag"

	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	"github.com/openshift-kni/debug-tools/pkg/irqs"
	softirqs "github.com/openshift-kni/debug-tools/pkg/irqs/soft"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [cpulist]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	var procfsRoot = flag.StringP("procfs", "P", "/proc", "procfs mount point to use.")
	var checkEffective = flag.BoolP("effective-affinity", "E", false, "check effective affinity.")
	var checkSoftirqs = flag.BoolP("softirqs", "S", false, "check softirqs counters.")
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

	var intersections int

	if *checkSoftirqs {
		src, err := os.Open(filepath.Join(*procfsRoot, "softirqs"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading softirqs from %q: %v", *procfsRoot, err)
			os.Exit(1)
		}
		info, err := softirqs.ReadInfo(src)
		src.Close()

		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing softirqs from %q: %v", *procfsRoot, err)
			os.Exit(1)
		}

		intersections = dumpSoftirqInfo(info, isolCpus)
	} else {
		flags := uint(0)
		if *checkEffective {
			flags |= irqs.EffectiveAffinity
		}

		irqInfos, err := irqs.ReadInfo(*procfsRoot, flags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing irqs from %q: %v", *procfsRoot, err)
			os.Exit(1)
		}

		intersections = dumpIrqInfo(irqInfos, isolCpus)
	}

	if intersections > 0 {
		os.Exit(1)
	}
}

func dumpIrqInfo(infos []irqs.Info, cpus cpuset.CPUSet) int {
	overlaps := 0
	for _, irqInfo := range infos {
		cpus := irqInfo.CPUs.Intersection(cpus)
		if cpus.Size() != 0 {
			fmt.Printf("IRQ %3d [%24s]: can run on %v\n", irqInfo.IRQ, irqInfo.Source, cpus.ToSlice())
			overlaps++
		}
	}
	return overlaps
}

func dumpSoftirqInfo(info *softirqs.Info, cpus cpuset.CPUSet) int {
	overlaps := 0
	keys := softirqs.Names()
	for _, key := range keys {
		counters := info.Counters[key]
		cb := cpuset.NewBuilder()
		for idx, counter := range counters {
			if counter > 0 {
				cb.Add(idx)
			}
		}
		usedCPUs := cpus.Intersection(cb.Result())
		if usedCPUs.Size() != 0 {
			overlaps++
		}
		fmt.Printf("%8s = %s\n", key, usedCPUs.String())
	}
	return overlaps
}
