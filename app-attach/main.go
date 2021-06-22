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
	"context"
	"time"

	"github.com/edwarnicke/govpp/binapi/session"
	"github.com/edwarnicke/log"
	"github.com/edwarnicke/vpphelper"
)

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

	appAttachReply, aErr := c.AppAttach(ctx, &session.AppAttach{})
	if aErr != nil {
		log.Entry(ctx).Fatalln("ERROR: AppAttach failed:", aErr)
	}
	log.Entry(ctx).Infof("Application Attached")

	log.Entry(ctx).Infof("App Msg Queue: %v\n"+
		"VPP Control Msg Queue: %v\n"+
		"VPP Control Queue Msg Thread Index: %v\n"+
		"App Index: %v\n"+
		"No. of fds exchanged: %v\n"+
		"FD Flags: %v\n"+
		"Segment Size: %v\n"+
		"Segment Handle: %v\n", appAttachReply.AppMq, appAttachReply.VppCtrlMq,
		appAttachReply.VppCtrlMqThread, appAttachReply.AppIndex, appAttachReply.NFds,
		appAttachReply.FdFlags, appAttachReply.SegmentSize, appAttachReply.SegmentHandle)

	// _, dErr := c.AppWorkerAddDel(ctx, &session.AppWorkerAddDel{AppIndex: appAttachReply.AppIndex, IsAdd: false})
	// if dErr != nil {
	// 	log.Entry(ctx).Fatalln("ERROR: App Worker Deletion failed", dErr)
	// }
	// log.Entry(ctx).Infof("App Worker Deleted")
	//	Cancel the context governing vpp's lifecycle and wait for it to exit
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
