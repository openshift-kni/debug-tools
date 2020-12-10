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

package irqs

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

const (
	EffectiveAffinity = 1 << iota
)

type Info struct {
	Source string
	IRQ    int
	CPUs   cpuset.CPUSet
}

func ReadInfo(procfsRoot string, flags uint) ([]Info, error) {
	irqRoot := filepath.Join(procfsRoot, "irq")

	files, err := ioutil.ReadDir(irqRoot)
	if err != nil {
		return nil, err
	}

	var irqs []int
	for _, file := range files {
		irq, err := strconv.Atoi(file.Name())
		if err != nil {
			continue // just skip not-irq-looking dirs
		}
		irqs = append(irqs, irq)
	}

	sort.Ints(irqs)

	affinityListFile := "smp_affinity_list"
	if (flags & EffectiveAffinity) == EffectiveAffinity {
		affinityListFile = "effective_affinity_list"
	}

	irqInfos := make([]Info, len(irqs))
	for _, irq := range irqs {
		irqDir := filepath.Join(irqRoot, fmt.Sprintf("%d", irq))

		irqCpuList, err := ioutil.ReadFile(filepath.Join(irqDir, affinityListFile))
		if err != nil {
			return nil, err
		}

		irqCpus, err := cpuset.Parse(strings.TrimSpace(string(irqCpuList)))
		if err != nil {
			continue // keep running
		}

		irqInfos = append(irqInfos, Info{
			CPUs:   irqCpus,
			IRQ:    irq,
			Source: findSourceForIRQ(procfsRoot, irq),
		})
	}
	return irqInfos, nil
}

func findSourceForIRQ(procfs string, irq int) string {
	irqDir := filepath.Join(procfs, "irq", fmt.Sprintf("%d", irq))
	files, err := ioutil.ReadDir(irqDir)
	if err != nil {
		return "MISSING"
	}
	for _, file := range files {
		if file.IsDir() {
			return file.Name()
		}
	}
	return ""
}
