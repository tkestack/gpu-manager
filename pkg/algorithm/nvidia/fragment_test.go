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

	"tkestack.io/gpu-manager/pkg/device/nvidia"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

func TestFragment(t *testing.T) {
	flag.Parse()
	obj := nvidia.NewNvidiaTree(nil)
	tree, _ := obj.(*nvidia.NvidiaTree)

	testCase1 :=
		`    GPU0    GPU1    GPU2    GPU3    GPU4    GPU5
GPU0      X      PIX     PHB     PHB     SOC     SOC
GPU1     PIX      X      PHB     PHB     SOC     SOC
GPU2     PHB     PHB      X      PIX     SOC     SOC
GPU3     PHB     PHB     PIX      X      SOC     SOC
GPU4     SOC     SOC     SOC     SOC      X      PIX
GPU5     SOC     SOC     SOC     SOC     PIX      X
`
	tree.Init(testCase1)
	algo := NewFragmentMode(tree)

	expectCase1 := []string{
		"/dev/nvidia4", "/dev/nvidia5",
	}

	cores := int64(2 * nvidia.HundredCore)
	pass, should, but := examining(expectCase1, algo.Evaluate(cores, 0))
	if !pass {
		t.Fatalf("Evaluate function got wrong, should be %s, but %s", should, but)
	}

	tree.MarkOccupied(&nvidia.NvidiaNode{
		Meta: nvidia.DeviceMeta{
			MinorID: 4,
		},
	}, cores, 0)

	expectCase2 := []string{
		"/dev/nvidia5",
	}

	cores = int64(nvidia.HundredCore)
	pass, should, but = examining(expectCase2, algo.Evaluate(cores, 0))
	if !pass {
		t.Fatalf("Evaluate function got wrong, should be %s, but %s", should, but)
	}
}

func TestFragmentOnlyOne(t *testing.T) {
	flag.Parse()
	obj := nvidia.NewNvidiaTree(nil)
	tree, _ := obj.(*nvidia.NvidiaTree)

	testCase1 :=
		` GPU0
GPU0   x`

	tree.Init(testCase1)
	algo := NewFragmentMode(tree)

	expectCase1 := []string{
		"/dev/nvidia0",
	}

	cores := int64(nvidia.HundredCore)
	pass, should, but := examining(expectCase1, algo.Evaluate(cores, 0))
	if !pass {
		t.Fatalf("Evaluate function got wrong, should be %s, but %s", should, but)
	}
}
