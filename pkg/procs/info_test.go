package procs_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/openshift-kni/debug-tools/pkg/procs"
)

var nullLog = log.New(io.Discard, "", 0)

func TestEmpty(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("creating temp dir %v", err)
	}
	defer os.RemoveAll(dir) // clean up

	ph := procs.New(nullLog, dir)
	pidInfos, err := ph.ListAll()
	if err != nil {
		t.Errorf("ListAll(%s) failed: %v", dir, err)
	}
	if len(pidInfos) > 0 {
		t.Errorf("found unexpected entries: %v", pidInfos)
	}
}

func TestSingleProcSingleThread(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("creating temp dir %v", err)
	}
	defer os.RemoveAll(dir) // clean up

	if err := makeFakeTree(dir, map[int]fakeEntry{
		1: fakeEntry{
			attrs: fakeAttrs{
				"cmdline": "/usr/lib/systemd/systemd\x00--switched-root\x00--system\x00--deserialize",
			},
			tasks: map[int]fakeAttrs{
				1: fakeAttrs{
					"status": "Name:	systemd\nPid:	1\nCpus_allowed_list:	0-3\n",
				},
			},
		},
	}); err != nil {
		t.Fatalf("populating temp dir %v", err)
	}

	ph := procs.New(nullLog, dir)
	pidInfos, err := ph.ListAll()
	if err != nil {
		t.Errorf("ListAll(%s) failed: %v", dir, err)
	}
	if len(pidInfos) != 1 {
		t.Errorf("found unexpected entries: %v", pidInfos)
	}

	expected := procs.PIDInfo{
		Pid:  1,
		Name: "systemd",
		TIDs: map[int]procs.TIDInfo{
			1: procs.TIDInfo{
				Tid:      1,
				Name:     "systemd",
				Affinity: []int{0, 1, 2, 3},
			},
		},
	}
	got := pidInfos[1]
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("unexpected return value: got=%#v expected=%#v", got, expected)
	}
}

type fakeAttrs map[string]string

type fakeEntry struct {
	attrs fakeAttrs
	tasks map[int]fakeAttrs
}

func makeFakeTree(root string, entries map[int]fakeEntry) error {
	for pid, entry := range entries {
		baseDir := filepath.Join(root, fmt.Sprintf("%d", pid))
		if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
			return err
		}

		for name, entry := range entry.attrs {
			if err := os.WriteFile(filepath.Join(baseDir, name), []byte(entry), os.ModePerm); err != nil {
				return err
			}
		}

		for tid, attrs := range entry.tasks {
			taskDir := filepath.Join(baseDir, "task", fmt.Sprintf("%d", tid))
			if err := os.MkdirAll(taskDir, os.ModePerm); err != nil {
				return err
			}
			for name, entry := range attrs {
				if err := os.WriteFile(filepath.Join(taskDir, name), []byte(entry), os.ModePerm); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
