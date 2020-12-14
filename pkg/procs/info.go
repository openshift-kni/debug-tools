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

package procs

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

type TIDInfo struct {
	Tid      int
	Name     string
	Affinity []int
}

type PIDInfo struct {
	Pid  int
	Name string
	TIDs map[int]TIDInfo
}

func All(procfsRoot string) (map[int]PIDInfo, error) {
	infos := make(map[int]PIDInfo)
	pidEntries, err := ioutil.ReadDir(procfsRoot)
	if err != nil {
		return infos, err
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

		pidInfo, err := FromPID(procfsRoot, pid)
		if err != nil {
			// TODO log
			continue
		}

		infos[pid] = pidInfo
	}
	return infos, nil
}

func FromPID(procfsRoot string, pid int) (PIDInfo, error) {
	pidInfo := PIDInfo{
		Pid:  pid,
		TIDs: make(map[int]TIDInfo),
	}

	if procName, err := readProcessName(procfsRoot, pid); err == nil {
		// failures not critical
		pidInfo.Name = basename(procName)
	}

	tasksDir := filepath.Join(procfsRoot, fmt.Sprintf("%d", pid), "task")
	tidEntries, err := ioutil.ReadDir(tasksDir)
	if err != nil {
		// TODO log
		return pidInfo, err
	}

	for _, tidEntry := range tidEntries {
		// TODO: use x/sys/unix schedGetAffinity?
		info, err := parseProcStatus(filepath.Join(tasksDir, tidEntry.Name(), "status"))
		if err != nil {
			continue
		}
		pidInfo.TIDs[info.Tid] = info
	}

	return pidInfo, nil
}

func parseProcStatus(path string) (TIDInfo, error) {
	info := TIDInfo{}
	file, err := os.Open(path)
	if err != nil {
		return info, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name:") {
			items := strings.SplitN(line, ":", 2)
			info.Name = strings.TrimSpace(items[1])
		}
		if strings.HasPrefix(line, "Pid:") {
			items := strings.SplitN(line, ":", 2)
			pid, err := strconv.Atoi(strings.TrimSpace(items[1]))
			if err != nil {
				return info, err
			}
			// yep, the field is called "pid" even if we are scaning /proc/XXX/task/YYY/status
			info.Tid = pid
		}
		if strings.HasPrefix(line, "Cpus_allowed_list:") {
			items := strings.SplitN(line, ":", 2)
			cpuIDs, err := cpuset.Parse(strings.TrimSpace(items[1]))
			if err != nil {
				return info, err
			}
			info.Affinity = cpuIDs.ToSlice()
		}
	}

	return info, scanner.Err()
}

func readProcessName(procfsRoot string, pid int) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(procfsRoot, fmt.Sprintf("%d", pid), "cmdline"))
	if err != nil {
		return "", err
	}
	cmdline := string(data)
	// workaround
	if off := strings.Index(cmdline, " "); off > 0 {
		return cmdline[:off], nil
	}
	if off := strings.Index(cmdline, "\x00"); off > 0 {
		return cmdline[:off], nil
	}
	return cmdline, nil
}

func basename(path string) string {
	name := filepath.Base(path)
	if name == "." {
		return ""
	}
	return name
}
