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

/*need to comment below two structs to prevent declared but not used error
import (
	"sync"
//	"unsafe"
)

// SvmMsgQT struct
type SvmMsgQT struct{
  q SvmMsgQqueueT
  rings []*SvmMsgQringT
}
// SvmMsgQqueueT struct
type  SvmMsgQqueueT struct {
  Shr *SvmMsgQsharedQueueT
  EvtFd int
  lock sync.Mutex
}

// SvmMsgQsharedQueueT struct
type SvmMsgQsharedQueueT struct {
  PadForPthreadCondvar [88]uint8 // set aside for pthread and condvar. to be implemented in future.
  Head uint32
  Tail uint32
  CurSize uint32
  MaxSize uint32
  ElSize uint32
  Pad uint32
}
// ClibSpinLockT struct
type  ClibSpinLockT struct {
  Lock uint32
  Pad1  uint32
  Pad2  [7]uint64
}
// SvmMsgQringT struct
type  SvmMsgQringT struct {
  NItems uint32
  ElSize uint32
  Shr *SvmMsgQRingSharedT
}
// SvmMsgQRingSharedT struct
type SvmMsgQRingSharedT struct{
  CurSize uint32
  NItems uint32
  Head uint32
  Tail uint32
  ElSize uint32
}
// SvmMsgQSharedT struct
type SvmMsgQSharedT struct {
  NRings uint32
  Pad uint32
}
// SvmMsgQmsgT1 struct
type  SvmMsgQmsgT1 struct {
  RingIndex uint32
  EltIndex uint32
}
// SvmMsgQmsgT2 struct
type  SvmMsgQmsgT2 struct {
  AsU64 uint64
}

  // The below code has not been tested and is incomplete.
  // A translation from SVM_MSG_QATTACH C code is attempted.
  // Minor changes will be needed as per the bugs we get due to Go language rules.
  // Commented the below code to prevent to golanggci-lint from throwing errors

func SvmMsgQAttach(mq *SvmMsgQT, SmqBase interface{}){
  var Ring *SvmMsgQRingSharedT // pthread condvar not supported now
  var Smq *SvmMsgQSharedT
  var i, NRings, QSize, Offset uint32

  Smq = (*SvmMsgQSharedT) &SmqBase
  mq.q.Shr = Smq + unsafe.Sizeof(Smq) // ? q is a [0] variable
  mq.q.EvtFd = -1
  NRings = Smq.NRings
  var rings = make([]*SvmMsgQringT, NRings)
  mq.rings = rings
  QSize = unsafe.Sizeof(SvmMsgQsharedQueueT) + mq.q.Shr.MaxSize * unsafe.Sizeof(SvmMsgQmsgT1) // SvmMsgQmsgT1 or SvmMsgQmsgT2 both have same size
  Ring = (interface{}) ((*uint8)(Smq + unsafe.Sizeof(Smq)) + QSize)
  for i := 0; i < int(NRings); i++{
    mq.rings[i].NItems = Ring.NItems
    mq.rings[i].ElSize = Ring.ElSize
    mq.rings[i].Shr = Ring.Shr
    Offset = unsafe.Sizeof(*Ring) + (Ring.NItems * Ring.ElSize)
    Ring = (interface{}) ((*uint8) (Ring + Offset))
	}
} */
