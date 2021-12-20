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
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"testing"
	"time"

	"tkestack.io/gpu-manager/pkg/config"
	"tkestack.io/gpu-manager/pkg/device/nvidia"
	"tkestack.io/gpu-manager/pkg/services/allocator/cache"
	"tkestack.io/gpu-manager/pkg/services/response"
	"tkestack.io/gpu-manager/pkg/services/watchdog"
	"tkestack.io/gpu-manager/pkg/types"
	"tkestack.io/gpu-manager/pkg/utils"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type podRawInfo struct {
	Name       string
	UID        string
	Containers []containerRawInfo
	OwnerKind  string
}

type containerRawInfo struct {
	Name             string
	Cores            int
	Memory           int
	PredicateIndexes string
}

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

func TestAllocatorRecover(t *testing.T) {
	flag.Parse()
	//init tree
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
	for _, n := range tree.Leaves() {
		n.AllocatableMeta.Cores = nvidia.HundredCore
		n.AllocatableMeta.Memory = 1024 * 1024 * 1024
		n.Meta.TotalMemory = 1024 * 1024 * 1024
	}

	//init allocator
	k8sClient := fake.NewSimpleClientset()
	alloc := initAllocator(tree, k8sClient)
	alloc.initEvaluator(tree)

	testCase := []struct {
		PodName         string
		PodUID          string
		Device          string
		Phase           string
		ContainerStatus []v1.ContainerStatus
	}{
		{
			PodName: "pending",
			PodUID:  "pending-uid",
			Device:  "/dev/nvidia0",
			Phase:   string(v1.PodPending),
		},
		{
			PodName: "failed",
			PodUID:  "failed-uid",
			Device:  "/dev/nvidia1",
			Phase:   string(v1.PodFailed),
		},
		{
			PodName: "containerExited",
			PodUID:  "contaienrExited-uid",
			Device:  "/dev/nvidia2",
			ContainerStatus: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 1,
						},
					},
				},
			},
		},
		{
			PodName: "pending",
			PodUID:  "pending-uid",
			Device:  "/dev/nvidia0",
			Phase:   string(v1.PodPending),
			ContainerStatus: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 1,
						},
					},
				},
			},
		},
	}

	podCache := cache.NewAllocateCache()
	//prepare ContainerCreateConfig and create the test container
	for _, testCase := range testCase {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: testCase.PodName,
				UID:  k8stypes.UID(testCase.PodUID),
				Annotations: map[string]string{
					types.VCoreAnnotation:   "100",
					types.VMemoryAnnotation: "1",
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: testCase.PodName,
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								types.VCoreAnnotation:   resource.MustParse("100"),
								types.VMemoryAnnotation: resource.MustParse("1"),
							},
						},
					},
				},
			},
		}
		if len(testCase.Phase) > 0 {
			pod.Status.Phase = v1.PodPhase(testCase.Phase)
		}
		if len(testCase.ContainerStatus) > 0 {
			pod.DeletionTimestamp = &metav1.Time{
				Time: time.Now(),
			}
		}
		if podCache.GetCache(string(pod.UID)) == nil {
			k8sClient.CoreV1().Pods("test-ns").Create(pod)
			podCache.Insert(string(pod.UID), testCase.PodName, &cache.Info{
				Devices: []string{testCase.Device},
				Cores:   100,
				Memory:  1,
			})
		}
	}

	watchdog.NewPodCacheForTest(k8sClient)
	data, err := json.Marshal(podCache)
	if err != nil {
		t.Errorf("Failed to marshal allocatedPod due to %v", err)
		return
	}
	alloc.checkpointManager.Write(data)
	defer alloc.checkpointManager.Delete()

	//test allocator recoverInUsed()
	alloc.recoverInUsed()
	allocatedPods := alloc.allocatedPod.Pods()
	if len(allocatedPods) != 1 || allocatedPods[0] != testCase[0].PodUID {
		t.Fatalf("allocated pods wrong: %v", allocatedPods)
	}

	expectAvailable := len(alloc.tree.Leaves()) - 1
	if alloc.tree.Available() != expectAvailable || alloc.tree.Query(testCase[0].Device).AllocatableMeta.Cores != 0 {
		t.Fatalf("node available wrong")
	}
}

func TestAllocator(t *testing.T) {
	flag.Parse()
	//init tree
	obj := nvidia.NewNvidiaTree(nil)
	tree, _ := obj.(*nvidia.NvidiaTree)

	expectObj := nvidia.NewNvidiaTree(nil)
	expectTree, _ := expectObj.(*nvidia.NvidiaTree)

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
	for _, n := range tree.Leaves() {
		n.AllocatableMeta.Cores = nvidia.HundredCore
		n.AllocatableMeta.Memory = 1024 * 1024 * 1024
		n.Meta.TotalMemory = 1024 * 1024 * 1024
	}

	//build expect tree
	expectTree.Init(testCase1)
	for _, n := range expectTree.Leaves() {
		n.AllocatableMeta.Cores = nvidia.HundredCore
		n.AllocatableMeta.Memory = 1024 * 1024 * 1024
		n.Meta.TotalMemory = 1024 * 1024 * 1024
	}
	for i := 0; i < 6; i++ {
		node := expectTree.Query("/dev/nvidia" + strconv.Itoa(i))
		expectTree.MarkOccupied(node, 100, 1*types.MemoryBlockSize)
	}
	t.Logf("expectTree graph: %s", expectTree.PrintGraph())

	//init allocator k8sclient and watchdog
	k8sClient := fake.NewSimpleClientset()
	watchdog.NewPodCacheForTest(k8sClient)
	alloc := initAllocator(tree, k8sClient)
	alloc.initEvaluator(tree)

	//create and allocate pod1
	raw1 := podRawInfo{
		Name: "pod-1",
		UID:  "uid-1",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            300,
				Memory:           3,
				PredicateIndexes: "0,1,2",
			},
			{
				Name: "container-without-gpu",
			},
		},
	}
	resps, err := createAndAllocate(alloc, k8sClient, raw1)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw1.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod1: %s", tree.PrintGraph())

	//create and allocate pod2
	raw2 := podRawInfo{
		Name: "pod-2",
		UID:  "uid-2",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            100,
				Memory:           1,
				PredicateIndexes: "3",
			},
			{
				Name:             "container-1",
				Cores:            100,
				Memory:           1,
				PredicateIndexes: "4",
			},
		},
	}
	resps, err = createAndAllocate(alloc, k8sClient, raw2)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw2.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod2: %s", tree.PrintGraph())

	//create and allocate pod3
	raw3 := podRawInfo{
		Name: "pod-3",
		UID:  "uid-3",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            100,
				Memory:           1,
				PredicateIndexes: "5",
			},
		},
	}
	resps, err = createAndAllocate(alloc, k8sClient, raw3)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw3.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod3: %s", tree.PrintGraph())

	//delete pod2
	deleteOption := metav1.DeleteOptions{}
	k8sClient.CoreV1().Pods("test-ns").Delete(raw2.Name, &deleteOption)

	//wait for watchdog to sync cache
	time.Sleep(1 * time.Second)

	//create and allocate pod4
	raw4 := podRawInfo{
		Name: "pod-4",
		UID:  "uid-4",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            50,
				Memory:           1,
				PredicateIndexes: "3",
			},
		},
	}
	resps, err = createAndAllocate(alloc, k8sClient, raw4)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw4.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod4: %s", tree.PrintGraph())

	//create and allocate pod5
	raw5 := podRawInfo{
		Name: "pod-5",
		UID:  "uid-5",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            60,
				Memory:           1,
				PredicateIndexes: "4",
			},
		},
	}
	resps, err = createAndAllocate(alloc, k8sClient, raw5)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw5.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod5: %s", tree.PrintGraph())

	//create and allocate pod6
	raw6 := podRawInfo{
		Name: "pod-6",
		UID:  "uid-6",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            50,
				Memory:           1,
				PredicateIndexes: "3",
			},
		},
	}
	resps, err = createAndAllocate(alloc, k8sClient, raw6)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw6.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod6: %s", tree.PrintGraph())

	//delete pod5
	deleteOption = metav1.DeleteOptions{}
	k8sClient.CoreV1().Pods("test-ns").Delete(raw5.Name, &deleteOption)
	time.Sleep(1 * time.Second)

	//create and allocate pod7
	raw7 := podRawInfo{
		Name: "pod-7",
		UID:  "uid-7",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            100,
				Memory:           1,
				PredicateIndexes: "4",
			},
		},
	}
	resps, err = createAndAllocate(alloc, k8sClient, raw7)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw7.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod7: %s", tree.PrintGraph())

	//compare tree with expectTree
	if !compareTree(tree, expectTree) {
		t.Fatalf("allocate test went wrong")
	}
}

func TestAllocateOneRepeatly(t *testing.T) {
	flag.Parse()
	//init tree
	obj := nvidia.NewNvidiaTree(nil)
	tree, _ := obj.(*nvidia.NvidiaTree)

	expectObj := nvidia.NewNvidiaTree(nil)
	expectTree, _ := expectObj.(*nvidia.NvidiaTree)

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
	for _, n := range tree.Leaves() {
		n.AllocatableMeta.Cores = nvidia.HundredCore
		n.AllocatableMeta.Memory = 1024 * 1024 * 1024
		n.Meta.TotalMemory = 1024 * 1024 * 1024
	}

	//build expect tree
	expectTree.Init(testCase1)
	for _, n := range expectTree.Leaves() {
		n.AllocatableMeta.Cores = nvidia.HundredCore
		n.AllocatableMeta.Memory = 1024 * 1024 * 1024
		n.Meta.TotalMemory = 1024 * 1024 * 1024
	}
	for i := 0; i < 6; i++ {
		node := expectTree.Query("/dev/nvidia" + strconv.Itoa(i))
		expectTree.MarkOccupied(node, 100, 1*types.MemoryBlockSize)
	}
	t.Logf("expectTree graph: %s", expectTree.PrintGraph())

	//init allocator k8sclient and watchdog
	k8sClient := fake.NewSimpleClientset()
	watchdog.NewPodCacheForTest(k8sClient)
	alloc := initAllocator(tree, k8sClient)
	alloc.initEvaluator(tree)

	//create and allocate pod1
	raw1 := podRawInfo{
		Name: "pod-1",
		UID:  "uid-1",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            600,
				Memory:           3,
				PredicateIndexes: "0,1,2,3,4,5",
			},
			{
				Name: "container-without-gpu",
			},
		},
	}
	resps, err := createAndAllocate(alloc, k8sClient, raw1)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw1.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod1: %s", tree.PrintGraph())

	resps, err = createAndAllocate(alloc, k8sClient, raw1)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw1.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod1: %s", tree.PrintGraph())

	//compare tree with expectTree
	if !compareTree(tree, expectTree) {
		t.Fatalf("allocate test went wrong")
	}
}

func TestAllocateOneFail(t *testing.T) {
	flag.Parse()
	//init tree
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
	for _, n := range tree.Leaves() {
		n.AllocatableMeta.Cores = nvidia.HundredCore
		n.AllocatableMeta.Memory = 1024 * 1024 * 1024
		n.Meta.TotalMemory = 1024 * 1024 * 1024
	}

	//init allocator k8sclient and watchdog
	k8sClient := fake.NewSimpleClientset()
	watchdog.NewPodCacheForTest(k8sClient)
	alloc := initAllocator(tree, k8sClient)
	alloc.initEvaluator(tree)

	//create and allocate pod1
	raw1 := podRawInfo{
		Name: "pod-1",
		UID:  "uid-1",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            5,
				Memory:           3,
				PredicateIndexes: "5",
			},
			{
				Name: "container-without-owner",
			},
		},
	}
	_, err := createAndAllocate(alloc, k8sClient, raw1)
	if err != nil {
		t.Logf("Failed to allocate for pod %s due to %+v", raw1.Name, err)
	}

	//create and allocate pod2
	raw2 := podRawInfo{
		Name:      "pod-2",
		UID:       "uid-2",
		OwnerKind: "Pod",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            5,
				Memory:           3,
				PredicateIndexes: "4",
			},
			{
				Name: "container-with-owner-pod",
			},
		},
	}
	_, err = createAndAllocate(alloc, k8sClient, raw2)
	if err != nil {
		t.Logf("Failed to allocate for pod %s due to %+v", raw2.Name, err)
	}

	//create and allocate pod2
	raw3 := podRawInfo{
		Name:      "pod-3",
		UID:       "uid-3",
		OwnerKind: "ReplicaSet",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            5,
				Memory:           3,
				PredicateIndexes: "3",
			},
			{
				Name: "container-with-owner-rs",
			},
		},
	}
	_, err = createAndAllocate(alloc, k8sClient, raw3)
	if err != nil {
		t.Logf("Failed to allocate for pod %s due to %+v", raw3.Name, err)
	}

	//wait for background check
	time.Sleep(3 * time.Second)

	_, err = k8sClient.CoreV1().Pods("test-ns").Get(raw1.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get pod1 due to %s", err.Error())
	}

	_, err = k8sClient.CoreV1().Pods("test-ns").Get(raw2.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get pod2 due to %s", err.Error())
	}

	_, err = k8sClient.CoreV1().Pods("test-ns").Get(raw3.Name, metav1.GetOptions{})
	if err == nil {
		t.Fatalf("pod3 should be deleted")
	}
}

func TestResponseManager(t *testing.T) {
	flag.Parse()
	//init tree
	obj := nvidia.NewNvidiaTree(nil)
	tree, _ := obj.(*nvidia.NvidiaTree)

	expectObj := nvidia.NewNvidiaTree(nil)
	expectTree, _ := expectObj.(*nvidia.NvidiaTree)

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
	for _, n := range tree.Leaves() {
		n.AllocatableMeta.Cores = nvidia.HundredCore
		n.AllocatableMeta.Memory = 1024 * 1024 * 1024
		n.Meta.TotalMemory = 1024 * 1024 * 1024
	}

	//build expect tree
	expectTree.Init(testCase1)
	for _, n := range expectTree.Leaves() {
		n.AllocatableMeta.Cores = nvidia.HundredCore
		n.AllocatableMeta.Memory = 1024 * 1024 * 1024
		n.Meta.TotalMemory = 1024 * 1024 * 1024
	}
	for i := 0; i < 6; i++ {
		node := expectTree.Query("/dev/nvidia" + strconv.Itoa(i))
		expectTree.MarkOccupied(node, 100, 1*types.MemoryBlockSize)
	}
	t.Logf("expectTree graph: %s", expectTree.PrintGraph())

	//init allocator k8sclient and watchdog
	k8sClient := fake.NewSimpleClientset()
	watchdog.NewPodCacheForTest(k8sClient)
	alloc := initAllocator(tree, k8sClient)
	alloc.initEvaluator(tree)

	//create and allocate pod1
	raw1 := podRawInfo{
		Name: "pod-1",
		UID:  "uid-1",
		Containers: []containerRawInfo{
			{
				Name:             "container-0",
				Cores:            600,
				Memory:           3,
				PredicateIndexes: "0,1,2,3,4,5",
			},
			{
				Name: "container-without-gpu",
			},
		},
	}
	resps, err := createAndAllocate(alloc, k8sClient, raw1)
	if err != nil {
		t.Errorf("Failed to allocate for pod %s due to %+v", raw1.Name, err)
	}
	t.Logf("resp: %+v", resps)
	t.Logf("tree graph after pod1: %s", tree.PrintGraph())

	if len(resps.ContainerResponses) != 1 {
		t.Errorf("expect only 1 container response, get %d", len(resps.ContainerResponses))
	}

	resp := resps.ContainerResponses[0]
	ctrlPath := utils.GetVirtualControllerMountPath(resp)
	if ctrlPath == "" {
		t.Errorf("controller path should not be empty")
	}

	expectedResp := alloc.responseManager.GetResp(raw1.UID, raw1.Containers[1].Name)
	if expectedResp != nil {
		t.Errorf("expected not found target response")
	}

	expectedResp = alloc.responseManager.GetResp(raw1.UID, raw1.Containers[0].Name)
	if expectedResp == nil {
		t.Errorf("expected found target response")
	}

	expectedCtrlPath := utils.GetVirtualControllerMountPath(expectedResp)
	if ctrlPath != expectedCtrlPath {
		t.Errorf("expected controller path to be %s, got %s,", ctrlPath, expectedCtrlPath)
	}

	k8sClient.CoreV1().Pods("test-ns").Delete(raw1.Name, &metav1.DeleteOptions{})
	time.Sleep(1 * time.Second)

	alloc.recycle()

	expectedResp = alloc.responseManager.GetResp(raw1.UID, raw1.Containers[0].Name)
	if expectedResp != nil {
		t.Errorf("expected not found target response")
	}
}

func createAndAllocate(alloc *NvidiaTopoAllocator, client kubernetes.Interface, raw podRawInfo) (*pluginapi.AllocateResponse, error) {
	var pod *v1.Pod
	pod, _ = client.CoreV1().Pods("test-ns").Get(raw.Name, metav1.GetOptions{})
	if pod == nil {
		pod = createPod(client, raw)
	}
	//wait for watchdog to sync cache
	time.Sleep(1 * time.Second)
	resps := &pluginapi.AllocateResponse{}
	for _, c := range pod.Spec.Containers {
		vcore := c.Resources.Limits[types.VCoreAnnotation]
		vmemory := c.Resources.Limits[types.VMemoryAnnotation]

		req := prepareContainerAllocateRequest(int(vcore.Value()), int(vmemory.Value()))
		alloc.recycle()
		resp, err := alloc.allocateOne(pod, &c, &req)
		if err != nil {
			pod.Status.Phase = v1.PodFailed
			pod.Status.Reason = types.UnexpectedAdmissionErrType
			pod.Status.Message = err.Error()
			_, _ = client.CoreV1().Pods("test-ns").UpdateStatus(pod)
			return resps, err
		}
		if resp != nil {
			resps.ContainerResponses = append(resps.ContainerResponses, resp)
		}
	}
	pod.Status.Phase = v1.PodRunning
	_, _ = client.CoreV1().Pods("test-ns").UpdateStatus(pod)
	_ = wait.Poll(time.Second, 5*time.Second, func() (bool, error) {
		newPod, _ := client.CoreV1().Pods("test-ns").Get(pod.Name, metav1.GetOptions{})
		if newPod.Status.Phase == v1.PodRunning {
			return true, nil
		}

		return false, nil
	})
	return resps, nil
}

func compareTree(tree1 *nvidia.NvidiaTree, tree2 *nvidia.NvidiaTree) bool {
	if !compareAllocatable(tree1, tree2) {
		return false
	}
	if !compareAvailble(tree1.Root(), tree2.Root()) {
		return false
	}
	return true
}

func compareAllocatable(tree1 *nvidia.NvidiaTree, tree2 *nvidia.NvidiaTree) bool {
	leaves1 := tree1.Leaves()
	nvidia.PrintSorter.Sort(leaves1)
	leaves2 := tree2.Leaves()
	nvidia.PrintSorter.Sort(leaves2)

	for i, n := range leaves1 {
		if n.AllocatableMeta.Cores != leaves2[i].AllocatableMeta.Cores {
			return false
		}
	}
	return true
}

func compareAvailble(node1 *nvidia.NvidiaNode, node2 *nvidia.NvidiaNode) bool {
	if node1.Available() != node2.Available() {
		return false
	}
	nvidia.PrintSorter.Sort(node1.Children)
	nvidia.PrintSorter.Sort(node2.Children)

	for i, n := range node1.Children {
		if !compareAvailble(n, node2.Children[i]) {
			return false
		}
	}
	return true
}

func createPod(client kubernetes.Interface, raw podRawInfo) *v1.Pod {
	containers := []v1.Container{}
	for _, c := range raw.Containers {
		container := v1.Container{
			Name: c.Name,
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					types.VCoreAnnotation:   resource.MustParse(fmt.Sprintf("%d", c.Cores)),
					types.VMemoryAnnotation: resource.MustParse(fmt.Sprintf("%d", c.Memory)),
				},
			},
		}
		containers = append(containers, container)
	}
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        raw.Name,
			UID:         k8stypes.UID(raw.UID),
			Annotations: make(map[string]string),
		},
		Spec: v1.PodSpec{
			Containers: containers,
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	}
	if raw.OwnerKind != "" {
		pod.OwnerReferences = []metav1.OwnerReference{{Kind: raw.OwnerKind, UID: "test-uid"}}
	}
	pod.Annotations[types.PredicateTimeAnnotation] = fmt.Sprintf("%d", time.Now().UnixNano())
	pod.Annotations[types.GPUAssigned] = "false"
	for i, c := range pod.Spec.Containers {
		if !utils.IsGPURequiredContainer(&c) {
			continue
		}
		pod.Annotations[types.PredicateGPUIndexPrefix+strconv.Itoa(i)] = raw.Containers[i].PredicateIndexes
	}
	pod, _ = client.CoreV1().Pods("test-ns").Create(pod)

	return pod
}

func prepareContainerAllocateRequest(cores int, memory int) (req pluginapi.ContainerAllocateRequest) {
	for i := 0; i < cores; i++ {
		req.DevicesIDs = append(req.DevicesIDs, types.VCoreAnnotation)
	}

	for i := 0; i < memory; i++ {
		req.DevicesIDs = append(req.DevicesIDs, types.VMemoryAnnotation)
	}
	return req
}

func initAllocator(tree *nvidia.NvidiaTree, client kubernetes.Interface) *NvidiaTopoAllocator {
	cfg := &config.Config{
		EnableShare:           true,
		VCudaRequestsQueue:    make(chan *types.VCudaRequest, 10),
		AllocationCheckPeriod: 2 * time.Second,
	}
	go func(cfg *config.Config) {
		for evt := range cfg.VCudaRequestsQueue {
			evt.Done <- nil
		}
	}(cfg)

	alloc := NewNvidiaTopoAllocatorForTest(cfg, tree, client, response.NewFakeResponseManager())
	return alloc.(*NvidiaTopoAllocator)
}
