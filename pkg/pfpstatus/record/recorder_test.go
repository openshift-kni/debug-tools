/*
 * Copyright 2025 Red Hat, Inc.
 *
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
 */

package record

import (
	"errors"
	"fmt"
	"testing"

	"github.com/k8stopologyawareschedwg/podfingerprint"
)

func TestNodeRecorderCreate(t *testing.T) {
	nr, err := NewNodeRecorder("test-node", WithCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nr.Len() != 0 {
		t.Errorf("non-empty recorder after creation")
	}
	x := nr.Cap()
	if x != 16 {
		t.Errorf("mismatching cap after creation: %v", x)
	}
}

func TestNodeRecorderCreateInvalid(t *testing.T) {
	_, err := NewNodeRecorder("test-node")
	if err == nil {
		t.Fatalf("unexpected success")
	}
	if !errors.Is(err, ErrInvalidCapacity) {
		t.Errorf("unexpected error: %v %T", err, err)
	}
}

func TestNodeRecorderPushMissingNodeName(t *testing.T) {
	nr, err := NewNodeRecorder("test-node", WithCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = nr.Push(podfingerprint.Status{
		NodeName: "",
	})
	if err == nil {
		t.Fatalf("unexpected success")
	}
	if !errors.Is(err, ErrMissingNode) {
		t.Fatalf("push: unexpected error: %v", err)
	}
}

func TestNodeRecorderPushMismatchingNodeName(t *testing.T) {
	nr, err := NewNodeRecorder("test-node", WithCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = nr.Push(podfingerprint.Status{
		NodeName: "FOOBAR",
	})
	if err == nil {
		t.Fatalf("unexpected success")
	}
	if !errors.Is(err, ErrMismatchingNode) {
		t.Fatalf("push: unexpected error: %v", err)
	}
}

func TestNodeRecorderPushBasic(t *testing.T) {
	nr, err := NewNodeRecorder("test-node", WithCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = nr.Push(podfingerprint.Status{
		NodeName: "test-node",
	})
	if err != nil {
		t.Fatalf("push: unexpected error: %v", err)
	}
	if nr.Len() != 1 {
		t.Errorf("empty recorder after push")
	}
}

func TestNodeRecorderSingleEntryPushEvict(t *testing.T) {
	nr, err := NewNodeRecorder("test-node", WithCapacity(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	num := 7 // random low enough number > 1
	for idx := 0; idx < num; idx++ {
		if err := nr.Push(podfingerprint.Status{
			NodeName:            "test-node",
			FingerprintExpected: fmt.Sprintf("pfp0v%d", idx),
		}); err != nil {
			t.Fatalf("push: unexpected error: %v", err)
		}
		sz := nr.Len()
		if sz != 1 {
			t.Fatalf("unexpected status count: %d", sz)
		}
	}
	sts := nr.Content()
	if len(sts) != 1 {
		t.Fatalf("unexpected content: %v", sts)
	}
	got := sts[0]
	exp := RecordedStatus{
		Status: podfingerprint.Status{
			NodeName:            "test-node",
			FingerprintExpected: "pfp0v6",
		},
	}
	if !exp.Equal(got) {
		t.Fatalf("unexpected content: got %#v exp %#v", got, exp)
	}
}

func TestNodeRecorderMultiEntryPushContentEvict(t *testing.T) {
	nr, err := NewNodeRecorder("test-node", WithCapacity(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	num := 7 // random low enough number > 1
	for idx := 0; idx < num; idx++ {
		if err := nr.Push(podfingerprint.Status{
			NodeName:            "test-node",
			FingerprintExpected: fmt.Sprintf("pfp0v%d", idx),
		}); err != nil {
			t.Fatalf("push: unexpected error: %v", err)
		}
		sz := nr.Len()
		if sz > 3 {
			t.Fatalf("unexpected status count: %d", sz)
		}
	}
	sts := nr.Content()
	if len(sts) != 3 {
		t.Fatalf("unexpected content: %v", sts)
	}
	for idx := 0; idx < 3; idx++ {
		got := sts[len(sts)-1-idx]
		exp := RecordedStatus{
			Status: podfingerprint.Status{
				NodeName:            "test-node",
				FingerprintExpected: fmt.Sprintf("pfp0v%d", num-1-idx),
			},
		}
		if !exp.Equal(got) {
			t.Fatalf("unexpected content: got %#v exp %#v", got, exp)
		}
	}
}

func TestRecorderCreate(t *testing.T) {
	rr, err := NewRecorder(WithMaxNodes(8), WithNodeCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := rr.Cap()
	if x != 16 {
		t.Errorf("mismatching cap after creation: %v", x)
	}
	n := rr.MaxNodes()
	if n != 8 {
		t.Errorf("mismatching maxNodes after creation: %v", n)
	}
	if rr.CountNodes() != 0 {
		t.Errorf("non-zero node count after creation")
	}
	if rr.Len() != 0 {
		t.Errorf("non-empty recorder after creation")
	}
}

func TestRecorderCreateInvalid(t *testing.T) {
	_, err := NewRecorder(WithMaxNodes(8), WithNodeCapacity(0))
	if err == nil {
		t.Fatalf("unexpected success")
	}
	if !errors.Is(err, ErrInvalidCapacity) {
		t.Errorf("unexpected error: %v %T", err, err)
	}

	_, err = NewRecorder(WithMaxNodes(0), WithNodeCapacity(32))
	if err == nil {
		t.Fatalf("unexpected success")
	}
	if !errors.Is(err, ErrInvalidNodeCount) {
		t.Errorf("unexpected error: %v %T", err, err)
	}

}

func TestRecorderPushMissingNodeName(t *testing.T) {
	rr, err := NewRecorder(WithMaxNodes(8), WithNodeCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = rr.Push(podfingerprint.Status{
		NodeName: "",
	})
	if err == nil {
		t.Fatalf("unexpected success")
	}
	if !errors.Is(err, ErrMissingNode) {
		t.Fatalf("push: unexpected error: %v", err)
	}
}

func TestRecorderPushTooManyNodes(t *testing.T) {
	rr, err := NewRecorder(WithMaxNodes(1), WithNodeCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := rr.Push(podfingerprint.Status{
		NodeName: "node-0",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rr.CountNodes() != 1 {
		t.Fatalf("unexpected node count")
	}
	err = rr.Push(podfingerprint.Status{
		NodeName: "node-1",
	})
	if err == nil {
		t.Fatalf("unexpected success")
	}
	if rr.CountNodes() != 1 {
		t.Fatalf("unexpected node count")
	}
	if rr.Len() != 1 {
		t.Fatalf("unexpected status count")
	}
	if rr.CountRecords("node-0") != 1 {
		t.Fatalf("unexpected status count for %q", "node-0")
	}
	if rr.CountRecords("node-1") != 0 {
		t.Fatalf("unexpected status count for %q", "node-1")
	}
	if !errors.Is(err, ErrTooManyNodes) {
		t.Fatalf("push: unexpected error: %v", err)
	}
}

func TestRecorderPushMultipleNodes(t *testing.T) {
	rr, err := NewRecorder(WithMaxNodes(8), WithNodeCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// gaps and dupes are intentional
	testNodes := []string{"node-0", "node-1", "node-3", "node-0"}
	for _, nodeName := range testNodes {
		if err := rr.Push(podfingerprint.Status{
			NodeName: nodeName,
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if rr.Len() != len(testNodes) {
		t.Fatalf("unexpected status count")
	}
	if rr.CountNodes() != 3 {
		t.Fatalf("unexpected node count")
	}
	if rr.CountRecords("node-0") != 2 {
		t.Fatalf("unexpected status count for %q", "node-0")
	}
}

func TestRecorderPushMultipleNodesContent(t *testing.T) {
	rr, err := NewRecorder(WithMaxNodes(8), WithNodeCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// gaps and dupes are intentional
	testNodes := []string{"node-0", "node-1", "node-3", "node-0"}
	for _, nodeName := range testNodes {
		if err := rr.Push(podfingerprint.Status{
			NodeName: nodeName,
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	sts := rr.Content()
	if len(sts) != 3 {
		t.Fatalf("unexpected node count")
	}
	if len(sts["node-0"]) != 2 {
		t.Fatalf("unexpected status count for %q", "node-0")
	}
	if len(sts["node-3"]) != 1 {
		t.Fatalf("unexpected status count for %q", "node-3")
	}
	exp := RecordedStatus{
		Status: podfingerprint.Status{
			NodeName: "node-3",
		},
	}
	if !exp.Equal(sts["node-3"][0]) {
		t.Fatalf("unexpected status for %q", "node-3")
	}

	got, ok := rr.ContentForNode("node-1")
	if !ok || len(got) != 1 {
		t.Fatalf("unexpected status for %q", "node-1")
	}
	exp = RecordedStatus{
		Status: podfingerprint.Status{
			NodeName: "node-1",
		},
	}
	if !exp.Equal(got[0]) {
		t.Fatalf("unexpected status for %q", "node-1")
	}
}

func TestRecorderPushMultipleNodesContentUnknownForNode(t *testing.T) {
	rr, err := NewRecorder(WithMaxNodes(8), WithNodeCapacity(16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// gaps and dupes are intentional
	testNodes := []string{"node-0", "node-1", "node-3", "node-0"}
	for _, nodeName := range testNodes {
		if err := rr.Push(podfingerprint.Status{
			NodeName: nodeName,
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	got, ok := rr.ContentForNode("node-99") // note node intentionally unknown
	if ok || len(got) != 0 {
		t.Fatalf("unexpected status count for %q", "node-99")
	}
}

func TestRecorderMultiPushSingleNode(t *testing.T) {
	rr, err := NewRecorder(WithMaxNodes(1), WithNodeCapacity(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 5; i++ {
		if err := rr.Push(podfingerprint.Status{
			NodeName: "node-0",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if rr.CountNodes() != 1 {
		t.Fatalf("unexpected node count")
	}
	if rr.CountRecords("node-0") != 5 {
		t.Fatalf("unexpected status count for %q", "node-0")
	}
}

func TestRecorderMultiPushMultipleNodes(t *testing.T) {
	rr, err := NewRecorder(WithMaxNodes(2), WithNodeCapacity(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 10; i++ {
		if err := rr.Push(podfingerprint.Status{
			NodeName: fmt.Sprintf("node-%d", i%2),
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if rr.CountNodes() != 2 {
		t.Fatalf("unexpected node count")
	}
	if rr.CountRecords("node-0") != 5 {
		t.Fatalf("unexpected status count for %q", "node-0")
	}
}
