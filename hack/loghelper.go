// loghelper.go is to be used by tests
// do not change its path!

package main

import (
	"bufio"
	"os"

	"github.com/openshift-kni/debug-tools/pkg/environ"
)

func main() {
	env := environ.New()
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		env.Log.Info("echo", "line", scan.Text())
	}
}
