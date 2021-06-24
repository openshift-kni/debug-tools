package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jaypipes/ghw/pkg/snapshot"
	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
)

const (
	snapprocName = "snapproc.tgz"
)

var _ = g.Describe("knit snapproc tests", func() {
	g.Context("With process snapshot", func() {
		g.It("Should be readable by cpulist", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "knit"),
				"snapproc",
			}
			fmt.Fprintf(g.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter
			_, err := cmd.Output()
			o.Expect(err).ToNot(o.HaveOccurred())
			defer os.Remove(snapprocName)

			snapPath, err := snapshot.Unpack(snapprocName)
			o.Expect(err).ToNot(o.HaveOccurred())
			defer os.RemoveAll(snapPath)

			cmdline = []string{
				filepath.Join(binariesPath, "knit"),
				"cpuaff",
				"-P",
				filepath.Join(snapPath, "proc"),
			}

			cmd = exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter
			// it's hard to predict the actual output, so we don't try yet.
			out, err := cmd.Output()
			o.Expect(err).ToNot(o.HaveOccurred())
			// but we expect _some_ output!
			o.Expect(out).ToNot(o.BeEmpty())

			cmdline = []string{
				filepath.Join(binariesPath, "cpulist"),
				"-P",
				filepath.Join(snapPath, "proc"),
				"-p",
				"1", // this is the pid most likely to be present
			}

			cmd = exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter
			// it's hard to predict the actual cpuset, so we don't try yet.
			_, err = cmd.Output()
			o.Expect(err).ToNot(o.HaveOccurred())
		})
	})
})
