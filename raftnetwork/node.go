// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raftnetwork

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/coreos/etcd/etcdserver/stats"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
	"golang.org/x/net/context"
)

var RaftNode *Node

type Node struct {
	raft.Node

	id        uint64
	stopc     chan struct{}
	pausec    chan bool
	storage   *raft.MemoryStorage
	mu        sync.Mutex // guards state
	state     raftpb.HardState
	transport *rafthttp.Transport

	httpstopc chan struct{} // signals http server to shutdown
}

func StartRaftNetwork(nodeID uint64, peers []string) *Node {
	var id uint64 = nodeID

	storage := raft.NewMemoryStorage()
	cfg := &raft.Config{
		ID:              id,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
	}

	raftpeers := make([]raft.Peer, len(peers))
	for i := range raftpeers {
		raftpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	rn := raft.StartNode(cfg, raftpeers)

	node := &Node{
		Node:      rn,
		id:        id,
		storage:   storage,
		pausec:    make(chan bool),
		httpstopc: make(chan struct{}),
	}
	node.start()

	// init status
	ss := &stats.ServerStats{}
	ss.Initialize()

	// start transport
	node.transport = &rafthttp.Transport{
		ID:          types.ID(node.id),
		ClusterID:   0x1000,
		Raft:        node,
		ServerStats: ss,
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(int(node.id))),
		ErrorC:      make(chan error),
	}

	node.transport.Start()
	for i, peer := range peers {
		if i+1 != int(node.id) {
			node.transport.AddPeer(types.ID(i+1), []string{peer})
		}
	}

	// start raft network
	go node.startServeRaft(peers)

	RaftNode = node
	return node
}

func (node *Node) startServeRaft(peers []string) {
	url, err := url.Parse(peers[node.id-1])
	if err != nil {
		log.Fatalf("raftexample: Failed parsing URL (%v)", err)
	}

	log.Println(url.Host, node.httpstopc)
	ln, err := NewRaftListener(url.Host, node.httpstopc)
	if err != nil {
		log.Fatalf("raftexample: Failed to listen rafthttp (%v)", err)
	}

	err = (&http.Server{Handler: node.transport.Handler()}).Serve(ln)
	select {
	case <-node.httpstopc:
	default:
		log.Fatalf("raftexample: Failed to serve rafthttp (%v)", err)
	}
	// close(node.httpdonec)
}

func (n *Node) start() {
	n.stopc = make(chan struct{})
	ticker := time.Tick(5 * time.Millisecond)

	go func() {
		for {
			select {
			case <-ticker:
				n.Tick()

			case rd := <-n.Ready():
				// if len(rd.Entries) > 0 {
				// 	log.Println(len(rd.Entries), string(rd.Entries[0].Data), rd.Entries[0].Term)
				// }
				// if len(rd.CommittedEntries) > 0 {
				// 	log.Println(len(rd.CommittedEntries), string(rd.CommittedEntries[0].Data), rd.CommittedEntries[0].Term)
				// }

				if !raft.IsEmptyHardState(rd.HardState) {
					n.mu.Lock()
					n.state = rd.HardState
					n.mu.Unlock()
					n.storage.SetHardState(n.state)
				}
				n.storage.Append(rd.Entries)
				n.transport.Send(rd.Messages)

				// time.Sleep(time.Millisecond)
				// TODO: make send async, more like real world...
				// for _, m := range rd.Messages {
				// 	n.iface.send(m)
				// }
				n.Advance()

			// case err := <-n.transport.ErrorC:
			// 	n.writeError(err)
			// 	return

			// case m := <-n.iface.recv():
			// 	// log.Printf("raft recv", m)
			// 	go n.Step(context.TODO(), m)

			case <-n.stopc:
				n.stop()
				return

				// case p := <-n.pausec:
				// 	log.Printf("raft pausec", p)
				// 	recvms := make([]raftpb.Message, 0)
				// 	for p {
				// 		select {
				// 		case m := <-n.iface.recv():
				// 			recvms = append(recvms, m)
				// 		case p = <-n.pausec:
				// 		}
				// 	}
				// 	// step all pending messages
				// 	for _, m := range recvms {
				// 		n.Step(context.TODO(), m)
				// 	}
			}
		}
	}()
}

// stop stops the node. stop a stopped node might panic.
// All in memory state of node is discarded.
// All stable MUST be unchanged.
func (node *Node) stop() {
	node.transport.Stop()
	node.Stop()

	// n.stopc <- struct{}{}
	// // wait for the shutdown
	// <-n.stopc
}

// restart restarts the node. restart a started node
// blocks and might affect the future stop operation.
func (n *Node) restart() {
	// wait for the shutdown
	<-n.stopc
	c := &raft.Config{
		ID:              n.id,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         n.storage,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
	}
	n.Node = raft.RestartNode(c)
	n.start()
}

// pause pauses the node.
// The paused node buffers the received messages and replies
// all of them when it resumes.
func (n *Node) pause() {
	n.pausec <- true
}

// resume resumes the paused node.
func (n *Node) resume() {
	n.pausec <- false
}

func (node *Node) IsIDRemoved(id uint64) bool                           { return false }
func (node *Node) ReportUnreachable(id uint64)                          {}
func (node *Node) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}

func (node *Node) Process(ctx context.Context, m raftpb.Message) error {
	return node.Step(ctx, m)
}
