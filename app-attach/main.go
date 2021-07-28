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
	"context"
	"time"

	"bytes"
	"encoding/binary"
	"net"
	"syscall"

	//	"github.com/justincormack/go-memfd"

	"github.com/edwarnicke/log"
	"github.com/edwarnicke/vpphelper"
	"github.com/harshgondaliya/govpp/binapi/session"
	"github.com/justincormack/go-memfd"
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
func main() {
	ctx, cancel1 := context.WithCancel(context.Background())
	// Connect to VPP with a 1 second timeout
	connectCtx, cancel2 := context.WithTimeout(ctx, time.Second)
	conn, vppErrCh := vpphelper.StartAndDialContext(connectCtx, vpphelper.WithVppConfig(vppConfContents))
	exitOnErrCh(ctx, cancel1, vppErrCh)

	// Create a RPC client for the session api
	c := session.NewServiceClient(conn)
	_, sErr := c.SessionEnableDisable(ctx, &session.SessionEnableDisable{IsEnable: true})
	if sErr != nil {
		log.Entry(ctx).Fatalln("ERROR: Session Enable Failed:", sErr)
	}
	log.Entry(ctx).Infof("Session Enabled")

	_, aErr := c.AppNamespaceAddDel(ctx, &session.AppNamespaceAddDel{NamespaceID: "12"})
	if aErr != nil {
		log.Entry(ctx).Fatalln("ERROR: Adding App Namespace Failed", sErr)
	}
	log.Entry(ctx).Infof("Added App Namespace")

	socketAddr := "/var/run/vpp/app_ns_sockets/12"
	udsConn, dErr := net.Dial("unixpacket", socketAddr)
	if dErr != nil {
		log.Entry(ctx).Fatalln("Dial Error", dErr)
	}
	log.Entry(ctx).Infof("Connected to Unix Socket")
	defer func() {
		_ = udsConn.Close()
	}()
	msg := AppSapiMsgAttach{MsgType: ATTACH, Msg: AppAttachMsg{Name: [64]uint8{97, 112, 112, 97, 116, 116, 97, 99, 104}, Options: [18]uint64{98}}}
	encMsg, encErr := msg.MarshalBinary()
	if encErr != nil {
		log.Entry(ctx).Fatalln("Encoding Error", encErr)
	}
	log.Entry(ctx).Infof("Encoding Successful")

	writer := bufio.NewWriter(udsConn)
	_, writeErr := writer.Write(encMsg)
	if writeErr != nil {
		log.Entry(ctx).Fatalln("Error while writing encoded message over connection", writeErr)
	}
	_ = writer.Flush()
	log.Entry(ctx).Infof("Successfully written message over connection")

	oob := make([]byte, syscall.CmsgSpace(4*int(2)))
	buf := make([]byte, 300) // 300 is arbitrary here, we should figure out how to make a wiser choice
	n, oobn, _, _, readConnErr := udsConn.(interface {
		ReadMsgUnix(b, oob []byte) (n, oobn, flags int, addr *net.UnixAddr, err error)
	}).ReadMsgUnix(buf, oob)
	if readConnErr != nil {
		log.Entry(ctx).Fatalln("Error while reading message from the connection. ", readConnErr)
	}
	buf = buf[:n]

	var replyMsg AppSapiMsgAttachReply
	decErr := replyMsg.UnmarshalBinary(buf)
	if decErr != nil {
		log.Entry(ctx).Fatalln("Error while decoding data read from the connection", decErr)
	}
	log.Entry(ctx).Infof("Successfully decoded message")
	log.Entry(ctx).Infof("Application Attached\n")

	log.Entry(ctx).Infof("App Index: %v\n"+
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
		log.Entry(ctx).Fatalln("Error while parsing socket control message. ", parseCtlErr)
	}
	var fdList []int
	for i := range msgs {
		fds, parseRightsErr := syscall.ParseUnixRights(&msgs[i])
		fdList = append(fdList, fds...)
		if parseRightsErr != nil {
			log.Entry(ctx).Fatalln("Error while parsing rights. ", parseRightsErr)
		}
	}
	var memBuf1 []byte
	var memBuf2 []byte
	var memBufErr1 error
	var memBufErr2 error

	if replyMsg.Msg.FdFlags&fdFlagVppMqSegment > 0 {
		mfdPtr1, newMfdErr1 := memfd.New(uintptr(fdList[0]))
		if newMfdErr1 != nil {
			log.Entry(ctx).Fatalln("Error while creating memfd. ", newMfdErr1)
		}
		memBuf1, memBufErr1 = mfdPtr1.Map()
		if memBufErr1 != nil {
			log.Entry(ctx).Fatalln("Error while creating memfd. ", memBufErr1)
		}
	}
	if replyMsg.Msg.FdFlags&fdFlagMemfdSegment > 0 {
		mfdPtr2, newMfdErr2 := memfd.New(uintptr(fdList[1]))
		if newMfdErr2 != nil {
			log.Entry(ctx).Fatalln("Error while creating memfd. ", newMfdErr2)
		}
		memBuf2, memBufErr2 = mfdPtr2.Map()
		if memBufErr2 != nil {
			log.Entry(ctx).Fatalln("Error while creating memfd. ", memBufErr2)
		}
	}
	_ = memBuf1
	_ = memBuf2
	cancel1()
	cancel2()
	<-vppErrCh
}

func exitOnErrCh(ctx context.Context, cancel context.CancelFunc, errCh <-chan error) {
	// If we already have an error, log it and exit
	select {
	case err := <-errCh:
		log.Entry(ctx).Fatal(err)
	default:
	}
	// Otherwise wait for an error in the background to log and cancel
	go func(ctx context.Context, errCh <-chan error) {
		err := <-errCh
		log.Entry(ctx).Error(err)
		cancel()
	}(ctx, errCh)
}
