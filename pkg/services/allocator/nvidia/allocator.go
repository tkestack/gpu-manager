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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	nveval "tkestack.io/gpu-manager/pkg/algorithm/nvidia"
	"tkestack.io/gpu-manager/pkg/config"
	"tkestack.io/gpu-manager/pkg/device"
	nvtree "tkestack.io/gpu-manager/pkg/device/nvidia"
	"tkestack.io/gpu-manager/pkg/services/allocator"
	"tkestack.io/gpu-manager/pkg/services/allocator/cache"
	"tkestack.io/gpu-manager/pkg/services/allocator/checkpoint"
	"tkestack.io/gpu-manager/pkg/services/response"
	"tkestack.io/gpu-manager/pkg/services/watchdog"
	"tkestack.io/gpu-manager/pkg/types"
	"tkestack.io/gpu-manager/pkg/utils"

	"golang.org/x/net/context"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	checkpointFileName = "gpumanager_internal_checkpoint"
)

func init() {
	allocator.Register("nvidia", NewNvidiaTopoAllocator)
	allocator.Register("nvidia_test", NewNvidiaTopoAllocatorForTest)
}

//NvidiaTopoAllocator is an allocator for Nvidia GPU
type NvidiaTopoAllocator struct {
	sync.Mutex

	tree         *nvtree.NvidiaTree
	allocatedPod *cache.PodCache

	config            *config.Config
	evaluators        map[string]Evaluator
	extraConfig       map[string]*config.ExtraConfig
	k8sClient         kubernetes.Interface
	unfinishedPod     *v1.Pod
	queue             workqueue.RateLimitingInterface
	stopChan          chan struct{}
	checkpointManager *checkpoint.Manager
	responseManager   response.Manager
}

const (
	ALLOCATE_SUCCESS = iota
	ALLOCATE_FAIL
	PREDICATE_MISSING
)

type allocateResult struct {
	pod     *v1.Pod
	result  int
	message string
	reason  string
	resChan chan struct{}
}

var (
	_           allocator.GPUTopoService = &NvidiaTopoAllocator{}
	waitTimeout                          = 10 * time.Second
)

//NewNvidiaTopoAllocator returns a new NvidiaTopoAllocator
func NewNvidiaTopoAllocator(config *config.Config,
	tree device.GPUTree,
	k8sClient kubernetes.Interface,
	responseManager response.Manager) allocator.GPUTopoService {

	_tree, _ := tree.(*nvtree.NvidiaTree)
	cm, err := checkpoint.NewManager(config.CheckpointPath, checkpointFileName)
	if err != nil {
		klog.Fatalf("Failed to create checkpoint manager due to %s", err.Error())
	}
	alloc := &NvidiaTopoAllocator{
		tree:              _tree,
		config:            config,
		evaluators:        make(map[string]Evaluator),
		allocatedPod:      cache.NewAllocateCache(),
		k8sClient:         k8sClient,
		queue:             workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		stopChan:          make(chan struct{}),
		checkpointManager: cm,
		responseManager:   responseManager,
	}

	// Load kernel module if it's not loaded
	alloc.loadModule()

	// Initialize evaluator
	alloc.initEvaluator(_tree)

	// Read extra config if it's given
	alloc.loadExtraConfig(config.ExtraConfigPath)

	// Process allocation results in another goroutine
	go wait.Until(alloc.runProcessResult, time.Second, alloc.stopChan)

	// Recover
	alloc.recoverInUsed()

	// Check allocation in another goroutine periodically
	go alloc.checkAllocationPeriodically(alloc.stopChan)

	return alloc
}

//NewNvidiaTopoAllocatorForTest returns a new NvidiaTopoAllocator
//with fake docker client, just for testing.
func NewNvidiaTopoAllocatorForTest(config *config.Config,
	tree device.GPUTree,
	k8sClient kubernetes.Interface,
	responseManager response.Manager) allocator.GPUTopoService {

	_tree, _ := tree.(*nvtree.NvidiaTree)
	cm, err := checkpoint.NewManager("/tmp", checkpointFileName)
	if err != nil {
		klog.Fatalf("Failed to create checkpoint manager due to %s", err.Error())
	}
	alloc := &NvidiaTopoAllocator{
		tree:              _tree,
		config:            config,
		evaluators:        make(map[string]Evaluator),
		allocatedPod:      cache.NewAllocateCache(),
		k8sClient:         k8sClient,
		stopChan:          make(chan struct{}),
		queue:             workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		checkpointManager: cm,
		responseManager:   responseManager,
	}

	// Initialize evaluator
	alloc.initEvaluator(_tree)

	// Check allocation in another goroutine periodically
	go alloc.checkAllocationPeriodically(alloc.stopChan)

	return alloc
}

func (ta *NvidiaTopoAllocator) runProcessResult() {
	for ta.processNextResult() {
	}
}

// #lizard forgives
func (ta *NvidiaTopoAllocator) recoverInUsed() {
	// Read and unmarshal data from checkpoint to allocatedPod before recover from docker
	ta.readCheckpoint()

	// Recover device tree by reading checkpoint file
	for uid, containerToInfo := range ta.allocatedPod.PodGPUMapping {
		for cName, cache := range containerToInfo {
			for _, dev := range cache.Devices {
				if utils.IsValidGPUPath(dev) {
					klog.V(2).Infof("Nvidia GPU %q is in use by container: %q", dev, cName)
					klog.V(2).Infof("Uid: %s, Name: %s, util: %d, memory: %d", uid, cName, cache.Cores, cache.Memory)

					id, _ := utils.GetGPUMinorID(dev)
					ta.tree.MarkOccupied(&nvtree.NvidiaNode{
						Meta: nvtree.DeviceMeta{
							MinorID: id,
						},
					}, cache.Cores, cache.Memory)
				}
			}
		}
	}

	ta.recycle()
	ta.writeCheckpoint()
	ta.checkAllocation()
}

func (ta *NvidiaTopoAllocator) checkAllocation() {
	klog.V(4).Infof("Checking allocation of pods on this node")
	pods, err := getPodsOnNode(ta.k8sClient, ta.config.Hostname, "")
	if err != nil {
		klog.Infof("Failed to get pods on node due to %v", err)
		return
	}

	for i, p := range pods {
		if !utils.IsGPURequiredPod(&p) {
			continue
		}
		switch p.Status.Phase {
		case v1.PodFailed, v1.PodPending:
			if utils.ShouldDelete(&pods[i]) {
				_ = ta.deletePodWithOwnerRef(&p)
			}
		case v1.PodRunning:
			annotaionMap, err := ta.getReadyAnnotations(&pods[i], true)
			if err != nil {
				klog.Infof("failed to get ready annotations for pod %s", p.UID)
				continue
			}
			pass := true
			for key, val := range annotaionMap {
				if v, ok := p.Annotations[key]; !ok || v != val {
					pass = false
					break
				}
			}
			if !pass {
				ar := &allocateResult{
					pod:     &pods[i],
					result:  ALLOCATE_SUCCESS,
					resChan: make(chan struct{}),
				}
				ta.queue.AddRateLimited(ar)
				<-ar.resChan
			}
		default:
			continue
		}
	}
}

func (ta *NvidiaTopoAllocator) checkAllocationPeriodically(quit chan struct{}) {
	ticker := time.NewTicker(ta.config.AllocationCheckPeriod)
	for {
		select {
		case <-ticker.C:
			ta.checkAllocation()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func (ta *NvidiaTopoAllocator) loadExtraConfig(path string) {
	if path != "" {
		klog.V(2).Infof("Load extra config from %s", path)

		f, err := os.Open(path)
		if err != nil {
			klog.Fatalf("Can not load extra config at %s, err %s", path, err)
		}

		defer f.Close()

		cfg := make(map[string]*config.ExtraConfig)
		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			klog.Fatalf("Can not unmarshal extra config, err %s", err)
		}

		ta.extraConfig = cfg
	}
}

func (ta *NvidiaTopoAllocator) initEvaluator(tree *nvtree.NvidiaTree) {
	ta.evaluators["link"] = nveval.NewLinkMode(tree)
	ta.evaluators["fragment"] = nveval.NewFragmentMode(tree)
	ta.evaluators["share"] = nveval.NewShareMode(tree)
}

func (ta *NvidiaTopoAllocator) loadModule() {
	if _, err := os.Stat(types.NvidiaCtlDevice); err != nil {
		if out, err := exec.Command("modprobe", "-va", "nvidia-uvm", "nvidia").CombinedOutput(); err != nil {
			klog.V(3).Infof("Running modprobe nvidia-uvm nvidia failed with message: %s, error: %v", out, err)
		}
	}

	if _, err := os.Stat(types.NvidiaUVMDevice); err != nil {
		if out, err := exec.Command("modprobe", "-va", "nvidia-uvm", "nvidia").CombinedOutput(); err != nil {
			klog.V(3).Infof("Running modprobe nvidia-uvm nvidia failed with message: %s, error: %v", out, err)
		}
	}
}

func (ta *NvidiaTopoAllocator) capacity() (devs []*pluginapi.Device) {
	var (
		gpuDevices, memoryDevices []*pluginapi.Device
		totalMemory               int64
	)

	nodes := ta.tree.Leaves()
	for i := range nodes {
		totalMemory += int64(nodes[i].Meta.TotalMemory)
	}

	totalCores := len(nodes) * nvtree.HundredCore
	gpuDevices = make([]*pluginapi.Device, totalCores)
	for i := 0; i < totalCores; i++ {
		gpuDevices[i] = &pluginapi.Device{
			ID:     fmt.Sprintf("%s-%d", types.VCoreAnnotation, i),
			Health: pluginapi.Healthy,
		}
	}

	totalMemoryBlocks := totalMemory / types.MemoryBlockSize
	memoryDevices = make([]*pluginapi.Device, totalMemoryBlocks)
	for i := int64(0); i < totalMemoryBlocks; i++ {
		memoryDevices[i] = &pluginapi.Device{
			ID:     fmt.Sprintf("%s-%d-%d", types.VMemoryAnnotation, types.MemoryBlockSize, i),
			Health: pluginapi.Healthy,
		}
	}

	devs = append(devs, gpuDevices...)
	devs = append(devs, memoryDevices...)

	return
}

// #lizard forgives
func (ta *NvidiaTopoAllocator) allocateOne(pod *v1.Pod, container *v1.Container, req *pluginapi.ContainerAllocateRequest) (*pluginapi.ContainerAllocateResponse, error) {
	var (
		nodes                       []*nvtree.NvidiaNode
		needCores, needMemoryBlocks int64
		predicateMissed             bool
		allocated                   bool
	)

	predicateMissed = !utils.IsGPUPredicatedPod(pod)
	singleNodeMemory := int64(ta.tree.Leaves()[0].Meta.TotalMemory)
	for _, v := range req.DevicesIDs {
		if strings.HasPrefix(v, types.VCoreAnnotation) {
			needCores++
		} else if strings.HasPrefix(v, types.VMemoryAnnotation) {
			needMemoryBlocks++
		}
	}

	if needCores == 0 && needMemoryBlocks == 0 {
		klog.Warningf("Zero request")
		return nil, nil
	}

	needMemory := needMemoryBlocks * types.MemoryBlockSize
	ta.tree.Update()
	shareMode := false

	podCache := ta.allocatedPod.GetCache(string(pod.UID))
	containerCache := &cache.Info{}
	if podCache != nil {
		if c, ok := podCache[container.Name]; ok {
			allocated = true
			containerCache = c
			klog.V(2).Infof("container %s of pod %s has already been allocated, the allcation will be skip", container.Name, pod.UID)
		}
	}
	klog.V(2).Infof("Tree graph: %s", ta.tree.PrintGraph())

	if allocated {
		klog.V(2).Infof("container %s of pod %s has already been allocated, get devices from cached instead", container.Name, pod.UID)
		for _, d := range containerCache.Devices {
			node := ta.tree.Query(d)
			if node != nil {
				nodes = append(nodes, node)
			}
		}
	} else {
		klog.V(2).Infof("Try allocate for %s(%s), vcore %d, vmemory %d", pod.UID, container.Name, needCores, needMemory)

		switch {
		case needCores > nvtree.HundredCore:
			eval, ok := ta.evaluators["link"]
			if !ok {
				return nil, fmt.Errorf("can not find evaluator link")
			}
			if needCores%nvtree.HundredCore > 0 {
				return nil, fmt.Errorf("cores are greater than %d, must be multiple of %d", nvtree.HundredCore, nvtree.HundredCore)
			}
			nodes = eval.Evaluate(needCores, 0)
		case needCores == nvtree.HundredCore:
			eval, ok := ta.evaluators["fragment"]
			if !ok {
				return nil, fmt.Errorf("can not find evaluator fragment")
			}
			nodes = eval.Evaluate(needCores, 0)
		default:
			if !ta.config.EnableShare {
				return nil, fmt.Errorf("share mode is not enabled")
			}
			if needCores == 0 || needMemory == 0 {
				return nil, fmt.Errorf("that cores or memory is zero is not permitted in share mode")
			}

			// evaluate in share mode
			shareMode = true
			eval, ok := ta.evaluators["share"]
			if !ok {
				return nil, fmt.Errorf("can not find evaluator share")
			}
			nodes = eval.Evaluate(needCores, needMemory)
			if len(nodes) == 0 {
				if shareMode && needMemory > singleNodeMemory {
					return nil, fmt.Errorf("request memory %d is larger than %d", needMemory, singleNodeMemory)
				}

				return nil, fmt.Errorf("no free node")
			}

			if !predicateMissed {
				// get predicate node by annotation
				containerIndex, err := utils.GetContainerIndexByName(pod, container.Name)
				if err != nil {
					return nil, err
				}
				var devStr string
				if idxStr, ok := pod.ObjectMeta.Annotations[types.PredicateGPUIndexPrefix+strconv.Itoa(containerIndex)]; ok {
					if _, err := strconv.Atoi(idxStr); err != nil {
						return nil, fmt.Errorf("predicate idx %s invalid for pod %s ", idxStr, pod.UID)
					}
					devStr = types.NvidiaDevicePrefix + idxStr
					if !utils.IsValidGPUPath(devStr) {
						return nil, fmt.Errorf("predicate idx %s invalid", devStr)
					}
				} else {
					return nil, fmt.Errorf("failed to find predicate idx for pod %s", pod.UID)
				}

				predicateNode := ta.tree.Query(devStr)
				if predicateNode == nil {
					return nil, fmt.Errorf("failed to get predicate node %s", devStr)
				}

				// check if we choose the same node as scheduler
				if predicateNode.MinorName() != nodes[0].MinorName() {
					return nil, fmt.Errorf("Nvidia node mismatch for pod %s(%s), pick up:%s  predicate: %s",
						pod.Name, container.Name, nodes[0].MinorName(), predicateNode.MinorName())
				}
			}
		}
	}

	if len(nodes) == 0 {
		if shareMode && needMemory > singleNodeMemory {
			return nil, fmt.Errorf("request memory %d is larger than %d", needMemory, singleNodeMemory)
		}

		return nil, fmt.Errorf("no free node")
	}

	ctntResp := &pluginapi.ContainerAllocateResponse{
		Envs:        make(map[string]string),
		Mounts:      make([]*pluginapi.Mount, 0),
		Devices:     make([]*pluginapi.DeviceSpec, 0),
		Annotations: make(map[string]string),
	}

	allocatedDevices := sets.NewString()
	deviceList := make([]string, 0)
	for _, n := range nodes {
		name := n.MinorName()
		klog.V(2).Infof("Allocate %s for %s(%s), Meta (%d:%d)", name, pod.UID, container.Name, n.Meta.ID, n.Meta.MinorID)

		ctntResp.Annotations[types.VCoreAnnotation] = fmt.Sprintf("%d", needCores)
		ctntResp.Annotations[types.VMemoryAnnotation] = fmt.Sprintf("%d", needMemory)

		ctntResp.Devices = append(ctntResp.Devices, &pluginapi.DeviceSpec{
			ContainerPath: name,
			HostPath:      name,
			Permissions:   "rwm",
		})
		deviceList = append(deviceList, n.Meta.UUID)

		if !allocated {
			ta.tree.MarkOccupied(n, needCores, needMemory)
		}
		allocatedDevices.Insert(name)
	}

	ctntResp.Annotations[types.VDeviceAnnotation] = vDeviceAnnotationStr(nodes)
	if !allocated {
		ta.allocatedPod.Insert(string(pod.UID), container.Name, &cache.Info{
			Devices: allocatedDevices.UnsortedList(),
			Cores:   needCores,
			Memory:  needMemory,
		})
	}

	// check if all containers of pod has been allocated; set unfinishedPod if not
	unfinished := false
	for _, c := range pod.Spec.Containers {
		if !utils.IsGPURequiredContainer(&c) {
			continue
		}
		podCache := ta.allocatedPod.GetCache(string(pod.UID))
		if podCache != nil {
			if _, ok := podCache[c.Name]; !ok {
				unfinished = true
				break
			}
		}
	}
	if unfinished {
		ta.unfinishedPod = pod
	} else {
		ta.unfinishedPod = nil
	}
	ta.writeCheckpoint()

	// Append control device
	ctntResp.Devices = append(ctntResp.Devices, &pluginapi.DeviceSpec{
		ContainerPath: types.NvidiaCtlDevice,
		HostPath:      types.NvidiaCtlDevice,
		Permissions:   "rwm",
	})

	ctntResp.Devices = append(ctntResp.Devices, &pluginapi.DeviceSpec{
		ContainerPath: types.NvidiaUVMDevice,
		HostPath:      types.NvidiaUVMDevice,
		Permissions:   "rwm",
	})

	// Append default device
	if cfg, found := ta.extraConfig["default"]; found {
		for _, dev := range cfg.Devices {
			ctntResp.Devices = append(ctntResp.Devices, &pluginapi.DeviceSpec{
				ContainerPath: dev,
				HostPath:      dev,
				Permissions:   "rwm",
			})
		}
	}

	// LD_LIBRARY_PATH
	ctntResp.Envs["LD_LIBRARY_PATH"] = "/usr/local/nvidia/lib64"
	for _, env := range container.Env {
		if env.Name == "compat32" && strings.ToLower(env.Value) == "true" {
			ctntResp.Envs["LD_LIBRARY_PATH"] = "/usr/local/nvidia/lib"
		}
	}

	// NVIDIA_VISIBLE_DEVICES
	ctntResp.Envs["NVIDIA_VISIBLE_DEVICES"] = strings.Join(deviceList, ",")

	if shareMode {
		ctntResp.Mounts = append(ctntResp.Mounts, &pluginapi.Mount{
			ContainerPath: "/usr/local/nvidia",
			HostPath:      types.DriverLibraryPath,
			ReadOnly:      true,
		})
	} else {
		ctntResp.Mounts = append(ctntResp.Mounts, &pluginapi.Mount{
			ContainerPath: "/usr/local/nvidia",
			HostPath:      types.DriverOriginLibraryPath,
			ReadOnly:      true,
		})
	}

	ctntResp.Mounts = append(ctntResp.Mounts, &pluginapi.Mount{
		ContainerPath: types.VCUDA_MOUNTPOINT,
		HostPath:      filepath.Join(ta.config.VirtualManagerPath, string(pod.UID)),
		ReadOnly:      true,
	})

	if predicateMissed {
		ar := &allocateResult{
			pod:     pod,
			result:  PREDICATE_MISSING,
			resChan: make(chan struct{}),
		}
		ta.queue.AddRateLimited(ar)
		<-ar.resChan
	}

	ta.responseManager.InsertResp(string(pod.UID), container.Name, ctntResp)

	return ctntResp, nil
}

func (ta *NvidiaTopoAllocator) requestForVCuda(podUID string) error {
	// Request for a independent directory for vcuda
	vcudaEvent := &types.VCudaRequest{
		PodUID: podUID,
		Done:   make(chan error, 1),
	}
	ta.config.VCudaRequestsQueue <- vcudaEvent
	return <-vcudaEvent.Done
}

func (ta *NvidiaTopoAllocator) recycle() {
	activePods := watchdog.GetActivePods()

	lastActivePodUids := sets.NewString()
	activePodUids := sets.NewString()
	for _, uid := range ta.allocatedPod.Pods() {
		lastActivePodUids.Insert(uid)
	}
	for uid := range activePods {
		activePodUids.Insert(uid)
	}

	podsToBeRemoved := lastActivePodUids.Difference(activePodUids)

	klog.V(5).Infof("Pods to be removed: %v", podsToBeRemoved.List())

	ta.freeGPU(podsToBeRemoved.List())
}

func (ta *NvidiaTopoAllocator) freeGPU(podUids []string) {
	for _, uid := range podUids {
		for contName, info := range ta.allocatedPod.GetCache(uid) {
			klog.V(2).Infof("Free %s(%s)", uid, contName)

			for _, devName := range info.Devices {
				id, _ := utils.GetGPUMinorID(devName)
				ta.tree.MarkFree(&nvtree.NvidiaNode{
					Meta: nvtree.DeviceMeta{
						MinorID: id,
					},
				}, info.Cores, info.Memory)
			}

			ta.responseManager.DeleteResp(uid, contName)
		}
		ta.allocatedPod.Delete(uid)
		if ta.unfinishedPod != nil && uid == string(ta.unfinishedPod.UID) {
			klog.V(2).Infof("unfinished pod %s was deleted, update cached reference to nil", uid)
			ta.unfinishedPod = nil
		}
	}
	ta.writeCheckpoint()
}

// #lizard forgives
//Allocate tries to allocate GPU node for each request
func (ta *NvidiaTopoAllocator) Allocate(_ context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	ta.Lock()
	defer ta.Unlock()

	var (
		reqCount           uint
		candidatePod       *v1.Pod
		candidateContainer *v1.Container
		found              bool
	)
	if len(reqs.ContainerRequests) < 1 {
		return nil, fmt.Errorf("empty container request")
	}

	// k8s send allocate request for one container at a time
	req := reqs.ContainerRequests[0]
	resps := &pluginapi.AllocateResponse{}
	reqCount = uint(len(req.DevicesIDs))

	klog.V(4).Infof("Request GPU device: %s", strings.Join(req.DevicesIDs, ","))

	ta.recycle()

	if ta.unfinishedPod != nil {
		candidatePod = ta.unfinishedPod
		cache := ta.allocatedPod.GetCache(string(candidatePod.UID))
		if cache == nil {
			msg := fmt.Sprintf("failed to find pod %s in cache", candidatePod.UID)
			klog.Infof(msg)
			return nil, fmt.Errorf(msg)
		}
		for i, c := range candidatePod.Spec.Containers {
			if _, ok := cache[c.Name]; ok {
				continue
			}

			if !utils.IsGPURequiredContainer(&c) {
				continue
			}

			if reqCount != utils.GetGPUResourceOfContainer(&candidatePod.Spec.Containers[i], types.VCoreAnnotation) {
				msg := fmt.Sprintf("allocation request mismatch for pod %s, reqs %v", candidatePod.UID, reqs)
				klog.Infof(msg)
				return nil, fmt.Errorf(msg)
			}
			candidateContainer = &candidatePod.Spec.Containers[i]
			found = true
			break
		}
	} else {
		pods, err := getCandidatePods(ta.k8sClient, ta.config.Hostname)
		if err != nil {
			msg := fmt.Sprintf("Failed to find candidate pods due to %v", err)
			klog.Infof(msg)
			return nil, fmt.Errorf(msg)
		}

		for _, pod := range pods {
			if found {
				break
			}
			for i, c := range pod.Spec.Containers {
				if !utils.IsGPURequiredContainer(&c) {
					continue
				}
				podCache := ta.allocatedPod.GetCache(string(pod.UID))
				if podCache != nil {
					if _, ok := podCache[c.Name]; ok {
						klog.Infof("container %s of pod %s has been allocate, continue to next", c.Name, pod.UID)
						continue
					}
				}
				if utils.GetGPUResourceOfContainer(&pod.Spec.Containers[i], types.VCoreAnnotation) == reqCount {
					klog.Infof("Found candidate Pod %s(%s) with device count %d", pod.UID, c.Name, reqCount)
					candidatePod = pod
					candidateContainer = &pod.Spec.Containers[i]
					found = true
					break
				}
			}
		}
	}

	if found {
		// get vmemory info from container spec
		vmemory := utils.GetGPUResourceOfContainer(candidateContainer, types.VMemoryAnnotation)
		for i := 0; i < int(vmemory); i++ {
			req.DevicesIDs = append(req.DevicesIDs, types.VMemoryAnnotation)
		}

		resp, err := ta.allocateOne(candidatePod, candidateContainer, req)
		if err != nil {
			klog.Errorf(err.Error())
			return nil, err
		}
		if resp != nil {
			resps.ContainerResponses = append(resps.ContainerResponses, resp)
		}
	} else {
		msg := fmt.Sprintf("candidate pod not found for request %v, allocation failed", reqs)
		klog.Infof(msg)
		return nil, fmt.Errorf(msg)
	}

	return resps, nil
}

//ListAndWatch is not implement
func (ta *NvidiaTopoAllocator) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	return fmt.Errorf("not implement")
}

//ListAndWatchWithResourceName send devices for request resource back to server
func (ta *NvidiaTopoAllocator) ListAndWatchWithResourceName(resourceName string, e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	devs := make([]*pluginapi.Device, 0)
	for _, dev := range ta.capacity() {
		if strings.HasPrefix(dev.ID, resourceName) {
			devs = append(devs, dev)
		}
	}

	s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})

	// We don't send unhealthy state
	for {
		time.Sleep(time.Second)
	}

	klog.V(2).Infof("ListAndWatch %s exit", resourceName)

	return nil
}

//GetDevicePluginOptions returns empty DevicePluginOptions
func (ta *NvidiaTopoAllocator) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{PreStartRequired: true}, nil
}

//PreStartContainer find the podUID by comparing request deviceids with deviceplugin
//checkpoint data, then checks the validation of allocation of the pod.
//Update pod annotation if check success, otherwise evict the pod.
func (ta *NvidiaTopoAllocator) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	ta.Lock()
	defer ta.Unlock()
	klog.V(2).Infof("get preStartContainer call from k8s, req: %+v", req)
	var (
		podUID        string
		containerName string
		vcore         int64
		vmemory       int64
		//devices       []string
	)

	// try to get podUID, containerName, vcore and vmemory from kubelet deviceplugin checkpoint file
	cp, err := utils.GetCheckpointData(ta.config.DevicePluginPath)
	if err != nil {
		msg := fmt.Sprintf("%s, failed to read from checkpoint file due to %v",
			types.PreStartContainerCheckErrMsg, err)
		klog.Infof(msg)
		return nil, fmt.Errorf(msg)
	}
	for _, entry := range cp.PodDeviceEntries {
		if entry.ResourceName == types.VCoreAnnotation &&
			utils.IsStringSliceEqual(req.DevicesIDs, entry.DeviceIDs) {
			podUID = entry.PodUID
			containerName = entry.ContainerName
			vcore = int64(len(entry.DeviceIDs))
			break
		}
	}

	for _, entry := range cp.PodDeviceEntries {
		if entry.PodUID == podUID &&
			entry.ContainerName == containerName &&
			entry.ResourceName == types.VMemoryAnnotation {
			vmemory = int64(len(entry.DeviceIDs))
			break
		}
	}

	if podUID == "" || containerName == "" {
		msg := fmt.Sprintf("%s, failed to get pod from deviceplugin checkpoint for PreStartContainer request %v",
			types.PreStartContainerCheckErrMsg, req)
		klog.Infof(msg)
		return nil, fmt.Errorf(msg)
	}
	pod, ok := watchdog.GetActivePods()[podUID]
	if !ok {
		msg := fmt.Sprintf("%s, failed to get pod %s in watchdog", types.PreStartContainerCheckErrMsg, podUID)
		klog.Infof(msg)
		return nil, fmt.Errorf(msg)
	}

	err = ta.preStartContainerCheck(podUID, containerName, vcore, vmemory)
	if err != nil {
		klog.Infof(err.Error())
		ta.queue.AddRateLimited(&allocateResult{
			pod:     pod,
			result:  ALLOCATE_FAIL,
			message: err.Error(),
			reason:  types.PreStartContainerCheckErrType,
		})
		return nil, err
	}

	// allocation check ok, request for VCuda to setup vGPU environment
	err = ta.requestForVCuda(podUID)
	if err != nil {
		msg := fmt.Sprintf("failed to setup VCuda for pod %s(%s) due to %v", podUID, containerName, err)
		klog.Infof(msg)
		return nil, fmt.Errorf(msg)
	}

	// prestart check pass, update pod annotation
	ar := &allocateResult{
		pod:     pod,
		result:  ALLOCATE_SUCCESS,
		resChan: make(chan struct{}),
	}
	ta.queue.AddRateLimited(ar)
	<-ar.resChan

	return &pluginapi.PreStartContainerResponse{}, nil
}

func (ta *NvidiaTopoAllocator) preStartContainerCheck(podUID string, containerName string, vcore int64, vmemory int64) error {
	cache := ta.allocatedPod.GetCache(podUID)
	if cache == nil {
		msg := fmt.Sprintf("%s, failed to get pod %s from allocatedPod cache",
			types.PreStartContainerCheckErrMsg, podUID)
		klog.Infof(msg)
		return fmt.Errorf(msg)
	}

	if c, ok := cache[containerName]; !ok {
		msg := fmt.Sprintf("%s, failed to get container %s of pod %s from allocatedPod cache",
			types.PreStartContainerCheckErrMsg, containerName, podUID)
		klog.Infof(msg)
		return fmt.Errorf(msg)
	} else if c.Memory != vmemory*types.MemoryBlockSize || c.Cores != vcore {
		// request and cache mismatch, evict the pod
		msg := fmt.Sprintf("%s, pod %s container %s requset mismatch from cache. req: vcore %d vmemory %d; cache: vcore %d vmemory %d",
			types.PreStartContainerCheckErrMsg, podUID, containerName, vcore, vmemory*types.MemoryBlockSize, c.Cores, c.Memory)
		klog.Infof(msg)
		return fmt.Errorf(msg)
	} else {
		devices := c.Devices
		if (vcore < nvtree.HundredCore && len(devices) != 1) ||
			(vcore >= nvtree.HundredCore && len(devices) != int(vcore/nvtree.HundredCore)) {
			msg := fmt.Sprintf("allocated devices mismatch, request for %d vcore, allocate %v", vcore, devices)
			klog.Infof(msg)
			return fmt.Errorf(msg)
		}
	}
	return nil
}

func (ta *NvidiaTopoAllocator) processNextResult() bool {
	// Wait until there is a new item in the working queue
	key, quit := ta.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer ta.queue.Done(key)

	result, ok := key.(*allocateResult)
	if !ok {
		klog.Infof("Failed to process result: %v, unable to translate to allocateResult", key)
		return true
	}
	// Invoke the method containing the business logic
	err := ta.processResult(result)
	// Handle the error if something went wrong during the execution of the business logic
	if err != nil {
		ta.queue.AddRateLimited(key)
		return true
	}

	ta.queue.Forget(key)
	return true
}

func (ta *NvidiaTopoAllocator) processResult(ar *allocateResult) error {
	switch ar.result {
	case ALLOCATE_SUCCESS:
		annotationMap, err := ta.getReadyAnnotations(ar.pod, true)
		if err != nil {
			msg := fmt.Sprintf("failed to get ready annotation of pod %s due to %s", ar.pod.UID, err.Error())
			klog.Infof(msg)
			return fmt.Errorf(msg)
		}
		err = patchPodWithAnnotations(ta.k8sClient, ar.pod, annotationMap)
		if err != nil {
			msg := fmt.Sprintf("add annotation for pod %s failed due to %s", ar.pod.UID, err.Error())
			klog.Infof(msg)
			return fmt.Errorf(msg)
		}
		close(ar.resChan)
	case ALLOCATE_FAIL:
		// free GPU devices that are already allocated to this pod
		ta.freeGPU([]string{string(ar.pod.UID)})

		ar.pod.Status = v1.PodStatus{
			Phase:   v1.PodFailed,
			Message: ar.message,
			Reason:  ar.reason,
		}
		ar.pod.Annotations = nil
		err := ta.updatePodStatus(ar.pod)
		if err != nil {
			msg := fmt.Sprintf("failed to set status of pod %s to PodFailed due to %s", ar.pod.UID, err.Error())
			klog.Infof(msg)
			return fmt.Errorf(msg)
		}
	case PREDICATE_MISSING:
		annotationMap, err := ta.getReadyAnnotations(ar.pod, false)
		err = patchPodWithAnnotations(ta.k8sClient, ar.pod, annotationMap)
		if err != nil {
			msg := fmt.Sprintf("add annotation for pod %s failed due to %s", ar.pod.UID, err.Error())
			klog.Infof(msg)
			return fmt.Errorf(msg)
		}
		close(ar.resChan)
	default:
		klog.Infof("unknown allocation result %d for pod %s", ar.result, ar.pod.UID)
	}
	return nil
}

func (ta *NvidiaTopoAllocator) getReadyAnnotations(pod *v1.Pod, assigned bool) (annotationMap map[string]string, err error) {
	//ta.Lock()
	//defer ta.Unlock()
	cache := ta.allocatedPod.GetCache(string(pod.UID))
	if cache == nil {
		msg := fmt.Sprintf("failed to get pod %s from allocatedPod cache", pod.UID)
		klog.Infof(msg)
		return nil, fmt.Errorf(msg)
	}

	annotationMap = make(map[string]string)
	for i, c := range pod.Spec.Containers {
		if !utils.IsGPURequiredContainer(&c) {
			continue
		}
		var devices []string
		containerCache, ok := cache[c.Name]
		if !ok {
			msg := fmt.Sprintf("failed to get container %s of pod %s from allocatedPod cache", c.Name, pod.UID)
			klog.Infof(msg)
			err = fmt.Errorf(msg)
			continue
		}

		devices = make([]string, len(containerCache.Devices))
		copy(devices, containerCache.Devices)
		for j, dev := range devices {
			strs := strings.Split(dev, types.NvidiaDevicePrefix)
			devices[j] = strs[len(strs)-1]
		}
		predicateIndexStr := strings.Join(devices, ",")
		annotationMap[types.PredicateGPUIndexPrefix+strconv.Itoa(i)] = predicateIndexStr
	}
	annotationMap[types.GPUAssigned] = strconv.FormatBool(assigned)

	return annotationMap, nil
}

func (ta *NvidiaTopoAllocator) updatePodStatus(pod *v1.Pod) error {
	klog.V(4).Infof("Try to update status of pod %s", pod.UID)

	err := wait.PollImmediate(time.Second, waitTimeout, func() (bool, error) {
		_, err := ta.k8sClient.CoreV1().Pods(pod.Namespace).UpdateStatus(pod)
		if err == nil {
			return true, nil
		}
		if utils.ShouldRetry(err) {
			klog.Infof("update status of pod %s failed due to %v, try again", pod.UID, err)
			newPod, err := ta.k8sClient.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			newPod.Status = pod.Status
			pod = newPod
			return false, nil
		}
		klog.V(4).Infof("Failed to update status of pod %s due to %v", pod.UID, err)
		return false, err
	})
	if err != nil {
		klog.Errorf("failed to update status of pod %s due to %v", pod.UID, err)
		return err
	}

	return nil
}

// delete pod if it is controlled by workloads like deployment, ignore naked pod
func (ta *NvidiaTopoAllocator) deletePodWithOwnerRef(pod *v1.Pod) error {
	// free GPU devices that are already allocated to this pod
	ta.freeGPU([]string{string(pod.UID)})

	if len(pod.OwnerReferences) > 0 {
		for _, ownerReference := range pod.OwnerReferences {
			// ignore pod if it is owned by another pod
			if ownerReference.Kind == pod.Kind {
				return nil
			}
		}
		// delete the pod
		klog.V(4).Infof("Try to delete pod %s", pod.UID)
		err := wait.PollImmediate(time.Second, waitTimeout, func() (bool, error) {
			err := ta.k8sClient.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{})
			if err == nil {
				return true, nil
			}
			if utils.ShouldRetry(err) {
				return false, nil
			}
			klog.V(4).Infof("Failed to delete pod %s due to %v", pod.UID, err)
			return false, err
		})
		if err != nil {
			klog.Errorf("failed to delete pod %s due to %v", pod.UID, err)
			return err
		}
	}

	return nil
}

func patchPodWithAnnotations(client kubernetes.Interface, pod *v1.Pod, annotationMap map[string]string) error {
	// update annotations by patching to the pod
	type patchMetadata struct {
		Annotations map[string]string `json:"annotations"`
	}
	type patchPod struct {
		Metadata patchMetadata `json:"metadata"`
	}
	payload := patchPod{
		Metadata: patchMetadata{
			Annotations: annotationMap,
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	err := wait.PollImmediate(time.Second, waitTimeout, func() (bool, error) {
		_, err := client.CoreV1().Pods(pod.Namespace).Patch(pod.Name, k8stypes.StrategicMergePatchType, payloadBytes)
		if err == nil {
			return true, nil
		}
		if utils.ShouldRetry(err) {
			return false, nil
		}

		return false, err
	})
	if err != nil {
		msg := fmt.Sprintf("failed to add annotation %v to pod %s due to %s", annotationMap, pod.UID, err.Error())
		klog.Infof(msg)
		return fmt.Errorf(msg)
	}
	return nil
}

func vDeviceAnnotationStr(nodes []*nvtree.NvidiaNode) string {
	str := make([]string, 0)
	for _, node := range nodes {
		str = append(str, node.MinorName())
	}

	return strings.Join(str, ",")
}

func getCandidatePods(client kubernetes.Interface, hostname string) ([]*v1.Pod, error) {
	candidatePods := []*v1.Pod{}
	allPods, err := getPodsOnNode(client, hostname, string(v1.PodPending))
	if err != nil {
		return candidatePods, err
	}
	for _, pod := range allPods {
		current := pod
		if utils.IsGPURequiredPod(&current) && !utils.IsGPUAssignedPod(&current) && !utils.ShouldDelete(&current) {
			candidatePods = append(candidatePods, &current)
		}
	}

	if klog.V(4) {
		for _, pod := range candidatePods {
			klog.Infof("candidate pod %s in ns %s with timestamp %d is found.",
				pod.Name,
				pod.Namespace,
				utils.GetPredicateTimeOfPod(pod))
		}
	}

	return OrderPodsdByPredicateTime(candidatePods), nil
}

func getPodsOnNode(client kubernetes.Interface, hostname string, phase string) ([]v1.Pod, error) {
	if len(hostname) == 0 {
		hostname, _ = os.Hostname()
	}
	var (
		selector fields.Selector
		pods     []v1.Pod
	)

	if phase != "" {
		selector = fields.SelectorFromSet(fields.Set{"spec.nodeName": hostname, "status.phase": phase})
	} else {
		selector = fields.SelectorFromSet(fields.Set{"spec.nodeName": hostname})
	}
	var (
		podList *v1.PodList
		err     error
	)

	err = wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
		podList, err = client.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
			FieldSelector: selector.String(),
			LabelSelector: labels.Everything().String(),
		})
		if err != nil {
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return pods, fmt.Errorf("failed to get Pods on node %s because: %v", hostname, err)
	}

	klog.V(9).Infof("all pods on this node: %v", podList.Items)
	for _, pod := range podList.Items {
		pods = append(pods, pod)
	}

	return pods, nil
}

// make the pod ordered by predicate time
func OrderPodsdByPredicateTime(pods []*v1.Pod) []*v1.Pod {
	newPodList := make(PodsOrderedByPredicateTime, 0, len(pods))
	for _, v := range pods {
		newPodList = append(newPodList, v)
	}
	sort.Sort(newPodList)
	return []*v1.Pod(newPodList)
}

type PodsOrderedByPredicateTime []*v1.Pod

func (pods PodsOrderedByPredicateTime) Len() int {
	return len(pods)
}

func (pods PodsOrderedByPredicateTime) Less(i, j int) bool {
	return utils.GetPredicateTimeOfPod(pods[i]) <= utils.GetPredicateTimeOfPod(pods[j])
}

func (pods PodsOrderedByPredicateTime) Swap(i, j int) {
	pods[i], pods[j] = pods[j], pods[i]
}

func (ta *NvidiaTopoAllocator) readCheckpoint() {
	data, err := ta.checkpointManager.Read()
	if err != nil {
		klog.Warningf("Failed to read from checkpoint due to %s", err.Error())
		return
	}
	err = json.Unmarshal(data, ta.allocatedPod)
	if err != nil {
		klog.Warningf("Failed to unmarshal data from checkpoint due to %s", err.Error())
	}
}

func (ta *NvidiaTopoAllocator) writeCheckpoint() {
	data, err := json.Marshal(ta.allocatedPod)
	if err != nil {
		klog.Warningf("Failed to marshal allocatedPod due to %s", err.Error())
		return
	}
	err = ta.checkpointManager.Write(data)
	if err != nil {
		klog.Warningf("Failed to write checkpoint due to %s", err.Error())
	}
}
