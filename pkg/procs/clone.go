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

package procs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// the proc layout we care about is more volatile than the ghw/snapshot package expects to deal with. So we need extra care.
func CloneInto(logger *log.Logger, procfsRoot, scratchDir string) error {
	baseDir := filepath.Join(scratchDir, "proc")

	var err error
	if err = os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return err
	}

	pidEntries, err := ioutil.ReadDir(procfsRoot)
	if err != nil {
		return err
	}
	for _, pidEntry := range pidEntries {
		if !pidEntry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(pidEntry.Name())
		if err != nil {
			// doesn't look like a pid
			continue
		}

		err = CloneEntry(logger, pid, procfsRoot, scratchDir)
		if err != nil {
			logger.Printf("cannot clone data for pid %d (skipped): %v", pid, err)
			continue
		}
	}
	return nil
}

func CloneEntry(logger *log.Logger, pid int, procfsRoot, scratchDir string) error {
	entryDir := filepath.Join(procfsRoot, fmt.Sprintf("%d", pid))
	var err error
	// keep the file open to (kinda-)lock the entry
	handle, err := os.Open(filepath.Join(entryDir, "cmdline"))
	if err != nil {
		return err
	}
	defer handle.Close()

	if err = cloneRunnableData(scratchDir, entryDir, "cmdline", "status"); err != nil {
		return err
	}

	taskRoot := filepath.Join(entryDir, "task")

	tidEntries, err := ioutil.ReadDir(taskRoot)
	if err != nil {
		return err
	}
	for _, tidEntry := range tidEntries {
		if !tidEntry.IsDir() {
			continue
		}
		tid, err := strconv.Atoi(tidEntry.Name())
		if err != nil {
			// doesn't look like a tid
			continue
		}

		err = cloneRunnableData(scratchDir, filepath.Join(taskRoot, tidEntry.Name()), "status")
		if err != nil {
			logger.Printf("cannot clone data for tid %d (skipped): %v", tid, err)
			continue
		}
	}

	return nil
}

func cloneRunnableData(scratchDir, baseDir string, items ...string) error {
	for _, item := range items {
		if err := os.MkdirAll(filepath.Join(scratchDir, baseDir), os.ModePerm); err != nil {
			return err
		}

		path := filepath.Join(baseDir, item)
		if err := copyFile(path, filepath.Join(scratchDir, path)); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(path, targetPath string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(targetPath, buf, os.ModePerm)
}
