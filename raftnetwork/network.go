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

// StartRaftNetwork start raft network
// func StartRaftNetwork() {
// 	peers := []raft.Peer{
// 		{1, nil},
// 		{2, nil},
// 		{3, nil},
// 	}
//
// 	storage := raft.NewMemoryStorage()
// 	cfg := &raft.Config{
// 		ID:              0x01,
// 		ElectionTick:    10,
// 		HeartbeatTick:   1,
// 		Storage:         storage,
// 		MaxSizePerMsg:   4096,
// 		MaxInflightMsgs: 256,
// 	}
// 	// Set peer list to the other nodes in the cluster.
// 	// Note that they need to be started separately as well.
// 	node := raft.StartNode(cfg, peers)
// 	log.Println(node)
//
// 	lead := node.Status().SoftState.Lead
// 	log.Println(lead)
// }
