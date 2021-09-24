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
	"bytes"
	"encoding/binary"

	log "github.com/sirupsen/logrus"
)

/* commented to avoid declared but not used error
// FsChunkVecLen constant
const (
	FsChunkVecLen = 11
)
*/

// FifoSegment struct
type FifoSegment struct {
	MemorySegment       *MemorySegment
	SegmentHeaderOffset uint64
	MqBytes             [][]byte // needs to mapped to fifo_segment underneath
}

// SegmentHeader struct
type SegmentHeader struct {
	NCachedBytes    uint64
	NActiveFifos    uint32
	NReservedBytes  uint32
	MaxLog2FifoSize uint32
	NSlices         uint8
	PctFirstAlloc   uint8
	NMqs            uint8
	Pad1            [41]byte
	ByteIndex       uint64
	MaxByteIndex    uint64
	StartByteIndex  uint64
	Pad2            [40]byte
	Slices          uint64
}

/* Commented below code to prevent declared but not used error
// FifoSegmentSlice struct
type FifoSegmentSlice struct
{
  Cacheline uint64
  FreeChunks[FsChunkVecLen] uint64
  FreeFifos uint64
  NflChunkBytes uint64
  VirtualMem uint64
  NumChunks [FsChunkVecLen]uint32
}

// NewFifoSegment function
func NewFifoSegment(memorySegment *MemorySegment) *FifoSegment {
	f := &FifoSegment{
    MemorySegment: memorySegment,
  }
  f.SegmentHeaderOffset, _ = binary.Uvarint(f.MemorySegment.MappedBytes[54:])

  sh1 := f.GetSegmentHeader()
  allocatedBytes := uint64(sh1.NReservedBytes) - sh1.StartByteIndex
  size :=  allocatedBytes/uint64(sh1.NMqs)

  sh2 := f.GetSegmentHeader()
  offset := f.SegmentHeaderOffset + sh2.StartByteIndex
  for i:= 0; i<int(sh2.NMqs); i++ {
    f.MqBytes = append(f.MqBytes, f.MemorySegment.MappedBytes[offset:offset+size])
  }
  return f
}
*/

// GetSegmentHeader function
func (f *FifoSegment) GetSegmentHeader() *SegmentHeader {
	segmentHeader := &SegmentHeader{}
	reader := bytes.NewReader(f.MemorySegment.MappedBytes[f.SegmentHeaderOffset:])
	err := binary.Read(reader, binary.LittleEndian, segmentHeader)
	if err != nil {
		log.Fatalf("ERROR: Binary reading to segment header %v", err)
	}
	return segmentHeader
}
