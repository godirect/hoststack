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
	"fmt"
	"net"
	"sync"

	"git.fd.io/govpp.git/api"
	log "github.com/sirupsen/logrus"
)

var doOnce sync.Once

// Connection struct
type Connection interface {
	api.Connection
	api.ChannelProvider
}

// Namespace struct
type Namespace struct {
	vppConn    Connection
	id         string
	udsConn    net.Conn
	attachment *Attachment
}

// NewNamespace function
func NewNamespace(conn Connection, id string) *Namespace {
	return &Namespace{ // should we pass udsConn too here
		vppConn: conn,
		id:      id,
	}
}

// Dial function
func (ns *Namespace) Dial() (net.Conn, error) {
	socketAddr := fmt.Sprintf("/var/run/vpp/app_ns_sockets/%v", ns.id)
	udsConn, dErr := net.Dial("unixpacket", socketAddr)
	if dErr != nil {
		log.Fatalf("Dial Error: %v", dErr)
	}
	log.Infof("Connected to Unix Socket")
	ns.udsConn = udsConn
	return udsConn, dErr
}

// Close function
func (ns *Namespace) Close() {
	_ = ns.udsConn.Close()
}

// Attach function
func (ns *Namespace) Attach() *Attachment {
	doOnce.Do(func() {
		ns.attachment = NewAttachment(ns, ns.udsConn)
	})
	return ns.attachment
}
