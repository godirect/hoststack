// Copyright (c) 2020 Cisco and/or its affiliates.
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
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// AppSapiMsgType type
type AppSapiMsgType int8

// ATTACH TYPE
const (
	ATTACH             AppSapiMsgType = iota + 1
	fdFlagVppMqSegment uint8          = 1
	fdFlagMemfdSegment uint8          = 2
)

// AppAttachMsg type
type AppAttachMsg struct {
	Name    [64]uint8
	Options [18]uint64
}

// AppAttachReplyMsg type
type AppAttachReplyMsg struct {
	Retval          int32
	AppIndex        uint32
	AppMq           uint64
	VppCtrlMq       uint64
	SegmentHandle   uint64
	APIClientHandle uint32
	VppCtrlMqThread uint8
	NFds            uint8
	FdFlags         uint8
}

// AppSapiMsgAttach type
type AppSapiMsgAttach struct {
	MsgType AppSapiMsgType
	Msg     AppAttachMsg
}

// AppSapiMsgAttachReply type
type AppSapiMsgAttachReply struct {
	MsgType AppSapiMsgType
	Msg     AppAttachReplyMsg
}

// MarshalBinary Function
func (msg *AppSapiMsgAttach) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, msg)
	return buf.Bytes(), err
}

// UnmarshalBinary Function
func (replyMsg *AppSapiMsgAttachReply) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, replyMsg)
	return err
}

// Attachment Struct
type Attachment struct {
	ns                 *Namespace
	appAttachReplyMsg  *AppAttachReplyMsg
	vppMqMemorySegment *MemorySegment
	workers            []*Worker
}

// NewAttachment function
func NewAttachment(ns *Namespace, udsConn io.Writer) *Attachment {
	attachment := &Attachment{
		ns: ns,
	}
	msg := AppSapiMsgAttach{MsgType: ATTACH, Msg: AppAttachMsg{Name: [64]uint8{97, 112, 112, 97, 116, 116, 97, 99, 104}, Options: [18]uint64{98}}}
	encMsg, encErr := msg.MarshalBinary()
	if encErr != nil {
		log.Fatalf("Encoding Error: %v", encErr)
	}
	log.Infof("Encoding Successful")

	writer := bufio.NewWriter(udsConn)
	_, writeErr := writer.Write(encMsg)
	if writeErr != nil {
		log.Fatalf("Error while writing encoded message over connection: %v", writeErr)
	}
	_ = writer.Flush()
	log.Infof("Successfully written message over connection")

	oob := make([]byte, syscall.CmsgSpace(4*int(2)))
	buf := make([]byte, 300) // 300 is arbitrary here, we should figure out how to make a wiser choice
	n, oobn, _, _, readConnErr := udsConn.(interface {
		ReadMsgUnix(b, oob []byte) (n, oobn, flags int, addr *net.UnixAddr, err error)
	}).ReadMsgUnix(buf, oob)
	if readConnErr != nil {
		log.Fatalf("Error while reading message from the connection: %v", readConnErr)
	}
	buf = buf[:n]

	var replyMsg AppSapiMsgAttachReply
	decErr := replyMsg.UnmarshalBinary(buf)
	if decErr != nil {
		log.Fatalf("Error while decoding data read from the connection: %v", decErr)
	}
	attachment.appAttachReplyMsg = &replyMsg.Msg
	log.Infof("Successfully decoded message")
	log.Infof("Application Attached\n")

	log.Infof("App Index: %v\n"+
		"App Message Queue: %v\n"+
		"VPP Control Message Queue: %v\n"+
		"Segment Handle: %v\n"+
		"API Client Handle: %v\n"+
		"VPP Control Message Queue Thread Index: %v\n"+
		"No. of fds exchanged: %v\n"+
		"FD Flags: %v\n", replyMsg.Msg.AppIndex, replyMsg.Msg.AppMq, replyMsg.Msg.VppCtrlMq, replyMsg.Msg.SegmentHandle,
		replyMsg.Msg.APIClientHandle, replyMsg.Msg.VppCtrlMqThread, replyMsg.Msg.NFds, replyMsg.Msg.FdFlags)

	msgs, parseCtlErr := syscall.ParseSocketControlMessage(oob[:oobn])
	if parseCtlErr != nil {
		log.Fatalf("Error while parsing socket control message: %v", parseCtlErr)
	}
	var fdList []int
	for i := range msgs {
		fds, parseRightsErr := syscall.ParseUnixRights(&msgs[i])
		fdList = append(fdList, fds...)
		if parseRightsErr != nil {
			log.Fatalf("Error while parsing rights: %v", parseRightsErr)
		}
	}

	if replyMsg.Msg.FdFlags&fdFlagVppMqSegment > 0 {
		attachment.vppMqMemorySegment = NewMemorySegment(fdList[0])
	}
	if replyMsg.Msg.FdFlags&fdFlagMemfdSegment > 0 {
		attachment.workers = append(attachment.workers, NewWorker(attachment, NewMemorySegment(fdList[1])))
	}
	return attachment
}
