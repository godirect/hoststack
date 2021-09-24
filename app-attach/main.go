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
	"context"
	"time"

	"github.com/edwarnicke/vpphelper"
	"github.com/harshgondaliya/govpp/binapi/session"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel1 := context.WithCancel(context.Background())
	// Connect to VPP with a 1 second timeout
	connectCtx, cancel2 := context.WithTimeout(ctx, time.Second)
	vppConn, vppErrCh := vpphelper.StartAndDialContext(connectCtx, vpphelper.WithVppConfig(vppConfContents))
	exitOnErrCh(cancel1, vppErrCh)

	// Create a RPC client for the session api
	c := session.NewServiceClient(vppConn)
	_, sErr := c.SessionEnableDisable(ctx, &session.SessionEnableDisable{IsEnable: true})
	if sErr != nil {
		log.Fatalf("ERROR: Session Enable Failed: %v", sErr)
	}
	log.Infof("Session Enabled")
	id := "12"
	_, aErr := c.AppNamespaceAddDel(ctx, &session.AppNamespaceAddDel{NamespaceID: id})
	if aErr != nil {
		log.Fatalf("ERROR: Adding App Namespace Failed %v", sErr)
	}
	log.Infof("Added App Namespace")
	ns := NewNamespace(vppConn, id)
	_, _ = ns.Dial()
	_ = ns.Attach()

	ns.Close()

	cancel1()
	cancel2()
	<-vppErrCh
}

func exitOnErrCh(cancel context.CancelFunc, errCh <-chan error) {
	// If we already have an error, log it and exit
	select {
	case err := <-errCh:
		log.Fatal(err)
	default:
	}
	// Otherwise wait for an error in the background to log and cancel
	go func(errCh <-chan error) {
		err := <-errCh
		log.Error(err)
		cancel()
	}(errCh)
}
