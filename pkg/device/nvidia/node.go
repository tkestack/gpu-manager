/*
 * Tencent is pleased to support the open source community by making TKEStack available.
 *
 * Copyright (C) 2012-2019 Tencent. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use
 * this file except in compliance with the License. You may obtain a copy of the
 * License at
 *
 * https://opensource.org/licenses/Apache-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OF ANY KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nvidia

import (
	"fmt"
	"math/bits"

	"k8s.io/klog"

	"tkestack.io/nvml"
)

//SchedulerCache contains allocatable resource of GPU
type SchedulerCache struct {
	Cores  int64
	Memory int64
}

//DeviceMeta contains metadata of GPU device
type DeviceMeta struct {
	ID          int
	MinorID     int
	UsedMemory  uint64
	TotalMemory uint64
	Pids        []uint
	BusId       string
	Utilization uint
	UUID        string
}

//NvidiaNode represents a node of Nvidia GPU
type NvidiaNode struct {
	Meta            DeviceMeta
	AllocatableMeta SchedulerCache

	Parent   *NvidiaNode
	Children []*NvidiaNode
	Mask     uint32

	pendingReset bool
	vchildren    map[int]*NvidiaNode
	ntype        nvml.GpuTopologyLevel
	tree         *NvidiaTree
}

var (
	/** test only */
	nodeIndex = 0
)

//NewNvidiaNode returns a new NvidiaNode
func NewNvidiaNode(t *NvidiaTree) *NvidiaNode {
	node := &NvidiaNode{
		vchildren: make(map[int]*NvidiaNode),
		ntype:     nvml.TOPOLOGY_UNKNOWN,
		tree:      t,
		Meta: DeviceMeta{
			ID: nodeIndex,
		},
	}

	nodeIndex++

	return node
}

func (n *NvidiaNode) setParent(p *NvidiaNode) {
	n.Parent = p
	p.vchildren[n.Meta.ID] = n
}

//MinorName returns MinorID of this NvidiaNode
func (n *NvidiaNode) MinorName() string {
	return fmt.Sprintf(NamePattern, n.Meta.MinorID)
}

//Type returns GpuTopologyLevel of this NvidiaNode
func (n *NvidiaNode) Type() int {
	return int(n.ntype)
}

//GetAvailableLeaves returns leaves of this NvidiaNode
//which available for allocating.
func (n *NvidiaNode) GetAvailableLeaves() []*NvidiaNode {
	var leaves []*NvidiaNode

	mask := n.Mask

	for mask != 0 {
		id := uint32(bits.TrailingZeros32(mask))
		klog.V(2).Infof("Pick up %d mask %b", id, n.tree.leaves[id].Mask)
		leaves = append(leaves, n.tree.leaves[id])
		mask ^= one << id
	}

	return leaves
}

//Available returns conut of available leaves
//of this NvidiaNode.
func (n *NvidiaNode) Available() int {
	return bits.OnesCount32(n.Mask)
}

func (n *NvidiaNode) String() string {
	switch n.ntype {
	case nvml.TOPOLOGY_INTERNAL:
		return fmt.Sprintf("GPU%d", n.Meta.ID)
	case nvml.TOPOLOGY_SINGLE:
		return "PIX"
	case nvml.TOPOLOGY_MULTIPLE:
		return "PXB"
	case nvml.TOPOLOGY_HOSTBRIDGE:
		return "PHB"
	case nvml.TOPOLOGY_CPU:
		return "CPU"
	case nvml.TOPOLOGY_SYSTEM:
		return "SYS"
	}

	return "ROOT"
}
