package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
	"github.com/openshift-kni/debug-tools/pkg/irqs"
)

var _ = g.Describe("knit IRQ watch tests", func() {

	var fixtureName = "dell_2_numa"
	var snapshotRoot string

	type cmdOutput struct {
		out []byte
		err error
	}

	g.Context("With isolated, reserved CPUs", func() {
		g.It("Produces the expected delta between consecutive irq samples", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "knit"),
				"-P", filepath.Join(snapshotRoot, "proc"),
				"-S", filepath.Join(snapshotRoot, "sys"),
				"-W", "3s",
				"-T", "1",
				"-J",
				"irqwatch",
			}
			fmt.Fprintf(g.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter

			c := make(chan cmdOutput)
			go func(c chan cmdOutput) {
				out, err := cmd.Output()
				c <- cmdOutput{
					out: out,
					err: err,
				}
			}(c)

			// while the irqwatch is running change the interrupts file and return the changes
			// this is very unlikley to suffer from race condition, since the irqwatch will wait 3 full seconds between each sample,
			// so modifyInterruptsFile function will finish it's job already
			delta, err := modifyInterruptsFile(filepath.Join(snapshotRoot, "proc"))
			o.Expect(err).ToNot(o.HaveOccurred())

			res := <-c
			o.Expect(res.err).ToNot(o.HaveOccurred())

			counters, err := extractCountersFromIRQWatchOutput(res.out)
			o.Expect(res.err).ToNot(o.HaveOccurred())

			ok := areCountersEqual(delta, counters)
			o.Expect(ok).To(o.BeTrue())
		})
	})

	g.BeforeEach(func() {
		snapshotRoot = snapshotBeforeEach(fixtureName, "sysinfo.tgz")
	})

	g.AfterEach(func() {
		snapshotAfterEach(snapshotRoot)
	})
})

func modifyInterruptsFile(procFs string) (irqs.Stats, error) {
	fakeLog := log.New(io.Discard, "", 0)

	ih := irqs.New(fakeLog, procFs)
	stats, err := ih.ReadStats()
	if err != nil {
		return nil, err
	}
	initialStats := stats.Clone()
	delta := make(irqs.Stats)
	rand.Seed(time.Now().UnixNano())

	// 50 random changes are good enough
	const numOfModifications int = 50
	// just a number that is big enough to make nice distribution of interrupts on random queues
	const maxInterrupts = 1000

	for i := 0; i < numOfModifications; i++ {
		numOfCpus := len(stats)

		// select random cpuid to change its counters
		cpuid := int(rand.Intn(numOfCpus))

		// select random IRQ in the seleced cpuid
		k := selRandKey(stats[cpuid])

		stats[cpuid][k] += uint64(rand.Intn(maxInterrupts))
		delta[cpuid] = make(irqs.Counter)
		delta[cpuid][k] = stats[cpuid][k] - initialStats[cpuid][k]
	}

	// add the random new stat values to the interruts file
	return delta, reWriteInterrupts(procFs, stats)
}

func reWriteInterrupts(procFs string, stats irqs.Stats) error {
	tmpf, err := os.CreateTemp(procFs, "interrupts")
	if err != nil {
		g.Fail(fmt.Sprintf("fail to create temp interrupts file %q", tmpf.Name()))
		return err
	}

	defer tmpf.Close()

	// Since it's a test we can assume no offlined cpu, so no holes in the slice (i.e. cpu_0 to cpu_n are presents)
	type cpuidToIrqValue []uint64

	// IRQ name maps to counter values for each cpuid
	// this way is makes it easier to be wrriten back into a file, line by line
	irqMap := make(map[string]cpuidToIrqValue)

	cpuids := getSortedCPUids(stats)

	for cpuid := range cpuids {
		for k, v := range stats[cpuid] {
			irqMap[k] = append(irqMap[k], v)
		}
	}

	tmpf.WriteString("\t")
	// write cpu ids on the first line
	for cpuid := range cpuids {
		cpuName := fmt.Sprintf("%11s", fmt.Sprintf("CPU%d", cpuid))
		tmpf.WriteString(cpuName)
	}
	tmpf.WriteString("\n")

	for irqName, counters := range irqMap {
		tmpf.WriteString(fmt.Sprintf("%4s:", irqName))

		for _, v := range counters {
			tmpf.WriteString(fmt.Sprintf("%11d", v))
		}
		tmpf.WriteString("\n")
	}

	// use this approach in order to avoid from other processes to access
	// the original interrupts file while modifing it
	interrruptsFile := filepath.Join(procFs, "interrupts")
	os.Rename(tmpf.Name(), interrruptsFile)
	if err != nil {
		g.Fail(fmt.Sprintf("fail to rename file %q to %q", tmpf.Name(), interrruptsFile))
		return err
	}

	return nil
}

func selRandKey(c irqs.Counter) string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys[rand.Intn(len(keys))]
}

func getSortedCPUids(s irqs.Stats) []int {
	keys := make([]int, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func extractCountersFromIRQWatchOutput(b []byte) (irqs.Stats, error) {
	type irqSummary struct {
		Elapsed  time.Duration `json:"elapsed"`
		Counters irqs.Stats    `json:"counters"`
	}

	irqS := irqSummary{}
	if err := json.Unmarshal(b, &irqS); err != nil {
		return nil, fmt.Errorf("Error unmarshalling b: %v", err)
	}

	return irqS.Counters, nil
}

func areCountersEqual(c1, c2 irqs.Stats) bool {
	return len(c1.Delta(c2)) == 0
}
