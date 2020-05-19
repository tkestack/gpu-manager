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
	"flag"
	"testing"

	"tkestack.io/gpu-manager/pkg/types"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

func TestTree(t *testing.T) {
	flag.Parse()
	testCase1 :=
		`    GPU0    GPU1    GPU2    GPU3    GPU4    GPU5
GPU0      X      PIX     PHB     PHB     SOC     SOC
GPU1     PIX      X      PHB     PHB     SOC     SOC
GPU2     PHB     PHB      X      PIX     SOC     SOC
GPU3     PHB     PHB     PIX      X      SOC     SOC
GPU4     SOC     SOC     SOC     SOC      X      PIX
GPU5     SOC     SOC     SOC     SOC     PIX      X
`
	testTree(t, testCase1, 6)

	testCase2 :=
		` GPU0
GPU0   x`
	testTree(t, testCase2, 1)
}

func testTree(t *testing.T, testCase string, nodeNum int) {
	//init tree
	obj := NewNvidiaTree(nil)
	tree, _ := obj.(*NvidiaTree)
	tree.Init(testCase)
	for _, n := range tree.Leaves() {
		n.AllocatableMeta.Cores = HundredCore
		n.AllocatableMeta.Memory = 1024
	}

	//test Leaves(), Total() and Available()
	leaves := tree.Leaves()
	if tree.Available() != nodeNum || len(leaves) != nodeNum || tree.Total() != nodeNum {
		t.Fatalf("available leaves number wrong")
	}

	//test Root() and GetAvailableLeaves()
	root := tree.Root()
	availableLeaves := root.GetAvailableLeaves()
	for i, l := range availableLeaves {
		if l != leaves[i] {
			t.Fatalf("get available leaves wrong")
		}
	}

	//test MarkOccupied() and MarkFree() with half core
	tree.MarkOccupied(leaves[0], 50, 1*types.MemoryBlockSize)
	if tree.Available() != (nodeNum - 1) {
		t.Fatalf("available leaves number wrong after MarkOccupied")
	}

	tree.MarkFree(leaves[0], 50, 1*types.MemoryBlockSize)
	if tree.Available() != nodeNum {
		t.Fatalf("available leaves number wrong after MarkFree")
	}

	//test MarkOccupied() and MarkFree() with one core
	tree.MarkOccupied(leaves[0], 100, 1*types.MemoryBlockSize)
	if tree.Available() != (nodeNum - 1) {
		t.Fatalf("available leaves number wrong after MarkOccupied")
	}

	tree.MarkFree(leaves[0], 100, 1*types.MemoryBlockSize)
	if tree.Available() != nodeNum {
		t.Fatalf("available leaves number wrong after MarkFree")
	}

	//test Query()
	if len(leaves) > 0 && tree.Query("/dev/nvidia0") != leaves[0] {
		t.Fatalf("method Query get wrong node")
	}
}
