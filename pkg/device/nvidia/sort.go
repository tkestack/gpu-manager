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
	"sort"

	"tkestack.io/gpu-manager/pkg/types"
)

//LessFunc represents funcion to compare two NvidiaNode
type LessFunc func(p1, p2 *NvidiaNode) bool

var (
	//ByType compares two NvidiaNode by GpuTopologyLevel
	ByType = func(p1, p2 *NvidiaNode) bool {
		return p1.Type() < p2.Type()
	}

	//ByAvailable compares two NvidiaNode by count of available leaves
	ByAvailable = func(p1, p2 *NvidiaNode) bool {
		return p1.Available() < p2.Available()
	}

	//ByID compares two NvidiaNode by ID
	ByID = func(p1, p2 *NvidiaNode) bool {
		return p1.Meta.ID < p2.Meta.ID
	}

	//ByMinorID compares two NvidiaNode by minor ID
	ByMinorID = func(p1, p2 *NvidiaNode) bool {
		return p1.Meta.MinorID < p2.Meta.MinorID
	}

	//ByMemory compares two NvidiaNode by memory already used
	ByMemory = func(p1, p2 *NvidiaNode) bool {
		return p1.Meta.UsedMemory < p2.Meta.UsedMemory
	}

	//ByPids compares two NvidiaNode by length of PIDs running on node
	ByPids = func(p1, p2 *NvidiaNode) bool {
		return len(p1.Meta.Pids) < len(p2.Meta.Pids)
	}

	//ByAllocatableCores compares two NvidiaNode by available cores
	ByAllocatableCores = func(p1, p2 *NvidiaNode) bool {
		return p1.AllocatableMeta.Cores < p2.AllocatableMeta.Cores
	}

	//ByAllocatableMemory compares two NvidiaNode by available memory
	ByAllocatableMemory = func(p1, p2 *NvidiaNode) bool {
		return p1.AllocatableMeta.Memory/types.MemoryBlockSize < p2.AllocatableMeta.Memory/types.MemoryBlockSize
	}

	//PrintSorter is used to sort nodes when printing them out
	PrintSorter = &printSort{
		less: []LessFunc{ByType, ByAvailable, ByMinorID},
	}
)

type printSort struct {
	data []*NvidiaNode
	less []LessFunc
}

func (p *printSort) Sort(d []*NvidiaNode) {
	p.data = d
	sort.Sort(p)
}

func (p *printSort) Len() int {
	return len(p.data)
}

func (p *printSort) Swap(i, j int) {
	p.data[i], p.data[j] = p.data[j], p.data[i]
}

func (p *printSort) Less(i, j int) bool {
	var k int

	for k = 0; k < len(p.less)-1; k++ {
		less := p.less[k]
		switch {
		case less(p.data[i], p.data[j]):
			return true
		case less(p.data[j], p.data[i]):
			return false
		}
	}

	return p.less[k](p.data[i], p.data[j])
}
