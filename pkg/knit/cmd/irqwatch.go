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
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/openshift-kni/debug-tools/pkg/irqs"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

type irqWatchOptions struct {
	period  string
	maxRuns int
	verbose int
}

func NewIRQWatchCommand(knitOpts *KnitOptions) *cobra.Command {
	opts := &irqWatchOptions{}
	irqWatch := &cobra.Command{
		Use:   "irqwatch",
		Short: "watch IRQ counters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return watchIRQs(cmd, knitOpts, opts, args)
		},
		Args: cobra.NoArgs,
	}
	irqWatch.Flags().IntVarP(&opts.maxRuns, "watch-times", "T", -1, "number of watch loops to perform, each every `watch-period`. Use -1 to run forever.")
	irqWatch.Flags().StringVarP(&opts.period, "watch-period", "W", "1s", "period to poll IRQ counters.")
	irqWatch.Flags().IntVarP(&opts.verbose, "verbose", "v", 1, "verbosiness amount.")
	return irqWatch
}

func watchIRQs(cmd *cobra.Command, knitOpts *KnitOptions, opts *irqWatchOptions, args []string) error {
	if opts.maxRuns == 0 {
		return nil
	}

	var err error
	period, err := time.ParseDuration(opts.period)
	if err != nil {
		return err
	}

	var initStats irqs.Stats
	var prevStats irqs.Stats
	var lastStats irqs.Stats

	ih := irqs.New(knitOpts.Log, knitOpts.ProcFSRoot)

	initTs := time.Now()
	initStats, err = ih.ReadStats()
	if err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	prevStats = initStats.Clone()
	ticker := time.NewTicker(period)
	reporter := NewIRQReporter(knitOpts.JsonOutput, opts.verbose, knitOpts.Cpus)

	done := false
	iterCount := 1
	for {
		select {
		case <-c:
			done = true
		case t := <-ticker.C:
			lastStats, err = ih.ReadStats()
			if err != nil {
				return err
			}
			reporter.Delta(t, prevStats, lastStats)
			prevStats = lastStats
		}

		if done {
			break
		}
		if opts.maxRuns > 0 && iterCount >= opts.maxRuns {
			break
		}
		iterCount++
	}

	reporter.Summary(initTs, initStats, lastStats)
	return nil
}

type IRQReporter interface {
	Delta(ts time.Time, prevStats, lastStats irqs.Stats)
	Summary(initTs time.Time, prevStats, lastStats irqs.Stats)
}

func NewIRQReporter(jsonOutput bool, verbose int, cpus cpuset.CPUSet) IRQReporter {
	if jsonOutput {
		return &irqReporterJSON{
			verbose: verbose,
			cpus:    cpus,
		}
	}
	return &irqReporterText{
		verbose: verbose,
		cpus:    cpus,
	}

}

type irqReporterText struct {
	verbose int
	cpus    cpuset.CPUSet
}

func (w *irqReporterText) Delta(ts time.Time, prevStats, lastStats irqs.Stats) {
	if w.verbose < 2 {
		return
	}
	delta := prevStats.Delta(lastStats)
	cpuids := w.cpus.ToSlice()
	for _, cpuid := range cpuids {
		counter, ok := delta[cpuid]
		if !ok {
			continue
		}
		for irqName, val := range counter {
			if val == 0 {
				continue
			}
			fmt.Printf("%v CPU=%d IRQ=%s +%d\n", ts, cpuid, irqName, val)
		}
	}
}

func (w *irqReporterText) Summary(initTs time.Time, prevStats, lastStats irqs.Stats) {
	if w.verbose < 1 {
		return
	}
	timeDelta := time.Now().Sub(initTs)
	delta := prevStats.Delta(lastStats)
	cpuids := w.cpus.ToSlice()

	fmt.Printf("\nIRQ summary on cpus %v after %v\n", w.cpus, timeDelta)
	for _, cpuid := range cpuids {
		counter, ok := delta[cpuid]
		if !ok {
			continue
		}
		for irqName, val := range counter {
			if val == 0 {
				continue
			}
			fmt.Printf("CPU=%d IRQ=%s +%d\n", cpuid, irqName, val)
		}
	}
}

type irqReporterJSON struct {
	verbose int
	cpus    cpuset.CPUSet
}

type irqDelta struct {
	Timestamp time.Time  `json:"timestamp"`
	Counters  irqs.Stats `json:"counters"`
}

func (w *irqReporterJSON) Delta(ts time.Time, prevStats, lastStats irqs.Stats) {
	if w.verbose < 2 {
		return
	}
	res := irqDelta{
		Timestamp: ts,
		Counters:  prevStats.Delta(lastStats).ForCPUs(w.cpus),
	}
	json.NewEncoder(os.Stdout).Encode(res)
}

type irqwatchDuration struct {
	d time.Duration
}

func (d irqwatchDuration) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.d.String())), nil
}

type irqSummary struct {
	Elapsed  irqwatchDuration `json:"elapsed"`
	Counters irqs.Stats       `json:"counters"`
}

func (w *irqReporterJSON) Summary(initTs time.Time, prevStats, lastStats irqs.Stats) {
	if w.verbose < 1 {
		return
	}
	res := irqSummary{
		Elapsed: irqwatchDuration{
			d: time.Now().Sub(initTs),
		},
		Counters: prevStats.Delta(lastStats).ForCPUs(w.cpus),
	}
	json.NewEncoder(os.Stdout).Encode(res)
}
