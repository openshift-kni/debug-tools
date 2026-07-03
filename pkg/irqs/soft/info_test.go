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

package soft_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	softirqs "github.com/openshift-kni/debug-tools/pkg/irqs/soft"
)

func TestNames(t *testing.T) {
	names := softirqs.Names()
	if names == nil || len(names) == 0 {
		t.Fatalf("missing names")
	}
}

var nullLog = log.New(io.Discard, "", 0)

func TestReadInfo(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("creating temp dir %v", err)
	}
	defer os.RemoveAll(rootDir) // clean up

	procDir := filepath.Join(rootDir, "proc")
	if err := os.Mkdir(procDir, 0755); err != nil {
		t.Fatalf("Mkdir(%s) failed: %v", procDir, err)
	}
	if err := os.WriteFile(filepath.Join(procDir, "softirqs"), []byte(fakeSoftirqs), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	ih := softirqs.New(nullLog, procDir)
	info, err := ih.ReadInfo()
	if err != nil {
		t.Errorf("ReadStats(%s) failed: %v", procDir, err)
	}

	var softirqTestCases = []struct {
		cpuIdx      int
		softirqName string
		value       uint64
	}{
		// some random non-zero values from the fakeSoftirqs below.
		// any non-zero value is fine, no special meaning.
		{0, "TIMER", 128764},
		{1, "TIMER", 200838},
		{2, "TIMER", 129415},
		{3, "TIMER", 129834},
		// now some zero values. Same criteria as above.
		{0, "IRQ_POLL", 0},
		{1, "IRQ_POLL", 0},
		{2, "IRQ_POLL", 0},
		{3, "IRQ_POLL", 0},
	}
	for _, tt := range softirqTestCases {
		t.Run(fmt.Sprintf("cpu %d softirq %q", tt.cpuIdx, tt.softirqName), func(t *testing.T) {
			v := info.Counters[tt.softirqName][tt.cpuIdx]
			if v != tt.value {
				t.Errorf("Counters mismatch got %v expected %v", v, tt.value)
			}
		})
	}
}

const fakeSoftirqs string = `                    CPU0       CPU1       CPU2       CPU3       
          HI:       3853     390251      75886       3513
       TIMER:     128764     200838     129415     129834
      NET_TX:         10          2       1282          3
      NET_RX:     162388         32          9         19
       BLOCK:       1083        482       1319        802
    IRQ_POLL:          0          0          0          0
     TASKLET:        626        118      33417         11
       SCHED:     424448     466847     406115     349634
     HRTIMER:          0          0          0          0
         RCU:     258340     286465     252164     250633`
