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

import "time"

type Option func(*Recorder)

func WithNodeCapacity(nodeCapacity int) Option {
	return func(rec *Recorder) {
		rec.nodeCapacity = nodeCapacity
	}
}

func WithNodeTimestamper(tsr func() time.Time) Option {
	return func(rec *Recorder) {
		rec.timestamper = tsr
	}
}

func WithMaxNodes(maxNodes int) Option {
	return func(rec *Recorder) {
		rec.maxNodes = maxNodes
	}
}

func WithPFPCoalescing(val bool) Option {
	return func(rr *Recorder) {
		rr.coalesceLast = val
	}
}

func WithMaxSizePerNode(maxSize int) Option {
	return func(rr *Recorder) {
		rr.maxSize = maxSize
	}
}

type NodeOption func(*NodeRecorder)

func WithCapacity(capacity int) NodeOption {
	return func(nr *NodeRecorder) {
		nr.capacity = capacity
	}
}

func WithTimestamper(tsr func() time.Time) NodeOption {
	return func(nr *NodeRecorder) {
		nr.timestamper = tsr
	}
}

func WithCoalescing(val bool) NodeOption {
	return func(nr *NodeRecorder) {
		nr.coalesceLast = val
	}
}

func WithMaxSize(maxSize int) NodeOption {
	return func(nr *NodeRecorder) {
		nr.maxSize = maxSize
	}
}
