// Copyright (c) 2021 Harsh Gondaliya.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/justincormack/go-memfd"
	log "github.com/sirupsen/logrus"
)

// MemorySegment struct
type MemorySegment struct {
	Memfd       *memfd.Memfd
	MappedBytes []byte
}

// NewMemorySegment function
func NewMemorySegment(fd int) *MemorySegment {
	mfdPtr, newMfdErr := memfd.New(uintptr(fd))
	if newMfdErr != nil {
		log.Fatalf("Error while creating memfd: %v", newMfdErr)
	}
	memSegment, memSegmentErr := mfdPtr.Map()
	if memSegmentErr != nil {
		log.Fatalf("Error while creating memfd: %v", memSegmentErr)
	}
	return &MemorySegment{
		Memfd:       mfdPtr,
		MappedBytes: memSegment,
	}
}
