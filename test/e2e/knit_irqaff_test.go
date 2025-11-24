package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("knit IRQ affinity tests", func() {

	var fixtureName = "dell_2_numa"

	var (
		dataDir      string
		snapshotRoot string
	)

	g.Context("With isolated, reserved CPUs", func() {
		g.It("Produces the expected affinity output", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "knit"),
				"-P", filepath.Join(snapshotRoot, "proc"),
				"-S", filepath.Join(snapshotRoot, "sys"),
				"-e",
				"-J",
				"irqaff",
			}
			fmt.Fprintf(g.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter

			out, err := cmd.Output()
			o.Expect(err).ToNot(o.HaveOccurred())
			refPath := filepath.Join(dataDir, "irqaff.json")
			fmt.Fprintf(g.GinkgoWriter, "reference data at: %q\n", refPath)

			expected, err := os.ReadFile(refPath)
			if err != nil {
				g.Fail(fmt.Sprintf("fail to read the irqaff reference data from %q", refPath))
			}

			diff, err := getJSONBlobsDiff(out, expected)
			if err != nil {
				g.Fail("fail to compare the irqaff reference")
			}
			o.Expect(diff).To(o.BeZero(), "unexpected JSON difference: %v", diff)
		})
	})

	g.BeforeEach(func() {
		dataDir = dataDirFor(fixtureName)
		snapshotRoot = snapshotBeforeEach(fixtureName, "sysinfo.tgz")
	})

	g.AfterEach(func() {
		snapshotAfterEach(snapshotRoot)
	})
})
