package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("cpulist", func() {
	g.Context("with arguments", func() {
		g.It("parses correctly commandline", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "cpulist"),
				"-c",
				"0-3,5,7-9",
			}
			fmt.Fprintf(g.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter

			out, err := cmd.Output()
			o.Expect(err).ToNot(o.HaveOccurred())

			expected := strings.Join([]string{"0", "1", "2", "3", "5", "7", "8", "9", ""}, "\n")

			o.Expect(string(out)).To(o.Equal(expected))
		})
		g.It("parses correctly stdin", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "cpulist"),
				"-f",
				"-",
			}
			fmt.Fprintf(g.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter
			cmd.Stdin = bytes.NewBufferString("0-5,7,9,12-15")

			out, err := cmd.Output()
			o.Expect(err).ToNot(o.HaveOccurred())

			expected := strings.Join([]string{"0", "1", "2", "3", "4", "5", "7", "9", "12", "13", "14", "15", ""}, "\n")

			o.Expect(string(out)).To(o.Equal(expected))
		})
	})

	g.Context("without arguments", func() {
		g.It("parses correctly /proc/self/status", func() {
			rootDir, err := os.MkdirTemp("", "test")
			if err != nil {
				g.Fail(fmt.Sprintf("creating temp dir %v", err))
			}
			defer os.RemoveAll(rootDir) // clean up

			procDir := filepath.Join(rootDir, "proc")
			procSelfTaskDir := filepath.Join(procDir, "self", "task", "self")
			if err := os.MkdirAll(procSelfTaskDir, 0755); err != nil {
				g.Fail(fmt.Sprintf("Mkdir(%s) failed: %v", procSelfTaskDir, err))
			}
			if err := os.WriteFile(filepath.Join(procSelfTaskDir, "status"), []byte(fakeSelfStatus), 0644); err != nil {
				g.Fail(fmt.Sprintf("WriteFile failed: %v", err))
			}

			cmdline := []string{
				filepath.Join(binariesPath, "cpulist"),
				"-P",
				procDir,
			}
			fmt.Fprintf(g.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter

			out, err := cmd.Output()
			o.Expect(err).ToNot(o.HaveOccurred())

			expected := strings.Join([]string{"0", "1", "2", "3", ""}, "\n")

			o.Expect(string(out)).To(o.Equal(expected))
		})

	})
})

const fakeSelfStatus string = `Name:	test
Umask:	0000
State:	S (sleeping)
Tgid:	1
Ngid:	0
Pid:	1
PPid:	0
TracerPid:	0
Uid:	0	0	0	0
Gid:	0	0	0	0
FDSize:	256
Groups:	 
NStgid:	1
NSpid:	1
NSpgid:	1
NSsid:	1
VmPeak:	  240812 kB
VmSize:	  175276 kB
VmLck:	       0 kB
VmPin:	       0 kB
VmHWM:	   17484 kB
VmRSS:	    9084 kB
RssAnon:	    3664 kB
RssFile:	    5420 kB
RssShmem:	       0 kB
VmData:	   22448 kB
VmStk:	     132 kB
VmExe:	     852 kB
VmLib:	   11292 kB
VmPTE:	     108 kB
VmSwap:	    2564 kB
HugetlbPages:	       0 kB
CoreDumping:	0
THP_enabled:	1
Threads:	1
SigQ:	1/63354
SigPnd:	0000000000000000
ShdPnd:	0000000000000000
SigBlk:	7be3c0fe28014a03
SigIgn:	0000000000001000
SigCgt:	00000001800004ec
CapInh:	0000000000000000
CapPrm:	000001ffffffffff
CapEff:	000001ffffffffff
CapBnd:	000001ffffffffff
CapAmb:	0000000000000000
NoNewPrivs:	0
Seccomp:	0
Seccomp_filters:	0
Cpus_allowed:	f
Cpus_allowed_list:	0-3
Mems_allowed:	00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000001
Mems_allowed_list:	0
voluntary_ctxt_switches:	6498
nonvoluntary_ctxt_switches:	1369`
