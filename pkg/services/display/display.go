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

package display

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	displayapi "tkestack.io/gpu-manager/pkg/api/runtime/display"
	"tkestack.io/gpu-manager/pkg/config"
	"tkestack.io/gpu-manager/pkg/device"
	nvtree "tkestack.io/gpu-manager/pkg/device/nvidia"
	"tkestack.io/gpu-manager/pkg/runtime"
	"tkestack.io/gpu-manager/pkg/services/watchdog"
	"tkestack.io/gpu-manager/pkg/types"
	"tkestack.io/gpu-manager/pkg/utils"
	"tkestack.io/gpu-manager/pkg/version"

	google_protobuf1 "github.com/golang/protobuf/ptypes/empty"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"tkestack.io/nvml"
)

//Display is used to show GPU device usage
type Display struct {
	sync.Mutex

	config                  *config.Config
	tree                    *nvtree.NvidiaTree
	containerRuntimeManager runtime.ContainerRuntimeInterface
}

var _ displayapi.GPUDisplayServer = &Display{}
var _ prometheus.Collector = &Display{}

//NewDisplay returns a new Display
func NewDisplay(config *config.Config, tree device.GPUTree, runtimeManager runtime.ContainerRuntimeInterface) *Display {
	_tree, _ := tree.(*nvtree.NvidiaTree)
	return &Display{
		tree:                    _tree,
		config:                  config,
		containerRuntimeManager: runtimeManager,
	}
}

//PrintGraph updates the tree and returns the result of tree.PrintGraph
func (disp *Display) PrintGraph(context.Context, *google_protobuf1.Empty) (*displayapi.GraphResponse, error) {
	disp.tree.Update()

	return &displayapi.GraphResponse{
		Graph: disp.tree.PrintGraph(),
	}, nil
}

//PrintUsages returns usage info getting from docker and watchdog
func (disp *Display) PrintUsages(context.Context, *google_protobuf1.Empty) (*displayapi.UsageResponse, error) {
	disp.Lock()
	defer disp.Unlock()

	activePods := watchdog.GetActivePods()
	displayResp := &displayapi.UsageResponse{
		Usage: make(map[string]*displayapi.ContainerStat),
	}

	for _, pod := range activePods {
		podUID := string(pod.UID)

		podUsage := disp.getPodUsage(pod)
		if len(podUsage) > 0 {
			displayResp.Usage[podUID] = &displayapi.ContainerStat{
				Cluster: pod.Annotations[types.ClusterNameAnnotation],
				Project: pod.Namespace,
				User:    getUserName(pod),
				Stat:    podUsage,
				Spec:    disp.getPodSpec(pod, podUsage),
			}
		}
	}

	if len(displayResp.Usage) > 0 {
		return displayResp, nil
	}

	return &displayapi.UsageResponse{}, nil
}

func (disp *Display) getPodSpec(pod *v1.Pod, devicesInfo map[string]*displayapi.Devices) map[string]*displayapi.Spec {
	podSpec := make(map[string]*displayapi.Spec)

	for _, ctnt := range pod.Spec.Containers {
		vcore := ctnt.Resources.Requests[types.VCoreAnnotation]
		vmemory := ctnt.Resources.Requests[types.VMemoryAnnotation]
		memBytes := vmemory.Value() * types.MemoryBlockSize

		spec := &displayapi.Spec{
			Gpu: float32(vcore.Value()) / 100,
			Mem: float32(memBytes >> 20),
		}

		if memBytes == 0 {
			var deviceMem int64
			if dev, ok := devicesInfo[ctnt.Name]; ok {
				for _, dev := range dev.Dev {
					deviceMem += int64(dev.DeviceMem)
				}
			}
			spec.Mem = float32(deviceMem)
		}

		podSpec[ctnt.Name] = spec
	}

	return podSpec
}

func (disp *Display) getPodUsage(pod *v1.Pod) map[string]*displayapi.Devices {
	podUsage := make(map[string]*displayapi.Devices)

	for _, stat := range pod.Status.ContainerStatuses {
		contName := stat.Name
		contID := strings.TrimPrefix(stat.ContainerID, fmt.Sprintf("%s://", disp.containerRuntimeManager.RuntimeName()))
		if len(contID) == 0 {
			continue
		}
		klog.V(4).Infof("Get container %s usage", contID)

		containerInfo, err := disp.containerRuntimeManager.InspectContainer(contID)
		if err != nil {
			klog.Warningf("can't find %s from docker", contID)
			continue
		}

		pidsInContainer, err := disp.containerRuntimeManager.GetPidsInContainers(contID)
		if err != nil {
			klog.Errorf("can't get pids form container %s, %v", contID, err)
			continue
		}
		_, _, deviceNames := utils.GetGPUData(containerInfo.Annotations)
		devicesUsage := make([]*displayapi.DeviceInfo, 0)
		for _, deviceName := range deviceNames {
			if utils.IsValidGPUPath(deviceName) {
				node := disp.tree.Query(deviceName)
				if usage := disp.getDeviceUsage(pidsInContainer, node.Meta.ID); usage != nil {
					usage.DeviceMem = float32(node.Meta.TotalMemory >> 20)
					devicesUsage = append(devicesUsage, usage)
				}
			}
		}

		if len(devicesUsage) > 0 {
			podUsage[contName] = &displayapi.Devices{
				Dev: devicesUsage,
			}
		}
	}

	return podUsage
}

//Version returns version of GPU manager
func (disp *Display) Version(context.Context, *google_protobuf1.Empty) (*displayapi.VersionResponse, error) {
	resp := &displayapi.VersionResponse{
		Version: version.Get().String(),
	}

	return resp, nil
}

func (disp *Display) getDeviceUsage(pidsInCont []int, deviceIdx int) *displayapi.DeviceInfo {
	nvml.Init()
	defer nvml.Shutdown()

	dev, err := nvml.DeviceGetHandleByIndex(uint(deviceIdx))
	if err != nil {
		klog.Warningf("can't find device %d, error %s", deviceIdx, err)
		return nil
	}

	processSamples, err := dev.DeviceGetProcessUtilization(1024, time.Second)
	if err != nil {
		klog.Warningf("can't get processes utilization from device %d, error %s", deviceIdx, err)
		return nil
	}

	processOnDevices, err := dev.DeviceGetComputeRunningProcesses(1024)
	if err != nil {
		klog.Warningf("can't get processes info from device %d, error %s", deviceIdx, err)
		return nil
	}

	busID, err := dev.DeviceGetPciInfo()
	if err != nil {
		klog.Warningf("can't get pci info from device %d, error %s", deviceIdx, err)
		return nil
	}

	sort.Slice(pidsInCont, func(i, j int) bool {
		return pidsInCont[i] < pidsInCont[j]
	})

	usedMemory := uint64(0)
	usedPids := make([]int32, 0)
	usedGPU := uint(0)
	for _, info := range processOnDevices {
		idx := sort.Search(len(pidsInCont), func(pivot int) bool {
			return pidsInCont[pivot] >= int(info.Pid)
		})

		if idx < len(pidsInCont) && pidsInCont[idx] == int(info.Pid) {
			usedPids = append(usedPids, int32(pidsInCont[idx]))
			usedMemory += info.UsedGPUMemory
		}
	}

	for _, sample := range processSamples {
		idx := sort.Search(len(pidsInCont), func(pivot int) bool {
			return pidsInCont[pivot] >= int(sample.Pid)
		})

		if idx < len(pidsInCont) && pidsInCont[idx] == int(sample.Pid) {
			usedGPU += sample.SmUtil
		}
	}

	return &displayapi.DeviceInfo{
		Id:      busID.BusID,
		CardIdx: fmt.Sprintf("%d", deviceIdx),
		Gpu:     float32(usedGPU),
		Mem:     float32(usedMemory >> 20),
		Pids:    usedPids,
	}
}

func getUserName(pod *v1.Pod) string {
	for _, env := range pod.Spec.Containers[0].Env {
		if env.Name == "SUBMITTER" {
			return env.Value
		}
	}

	return ""
}

type gpuUtilDesc struct{}
type gpuUtilSpecDesc struct{}
type gpuMemoryDesc struct{}
type gpuMemorySpecDesc struct{}

var (
	defaultMetricLabels   = []string{"pod_name", "namespace", "node", "container_name"}
	utilDescBuilder       = gpuUtilDesc{}
	utilSpecDescBuilder   = gpuUtilSpecDesc{}
	memoryDescBuilder     = gpuMemoryDesc{}
	memorySpecDescBuilder = gpuMemorySpecDesc{}
)

const (
	metricPodName = iota
	metricNamespace
	metricNodeName
	metricContainerName
)

func (gpuUtilDesc) getDescribeDesc() *prometheus.Desc {
	return prometheus.NewDesc("container_gpu_utilization", "gpu utilization", []string{"gpu"}, nil)
}

func (gpuUtilSpecDesc) getDescribeDesc() *prometheus.Desc {
	return prometheus.NewDesc("container_request_gpu_utilization", "request of gpu utilization", []string{"req_of_gpu"}, nil)
}

func (gpuUtilDesc) getMetricDesc() *prometheus.Desc {
	return prometheus.NewDesc("container_gpu_utilization", "gpu utilization", append(defaultMetricLabels, "gpu"), nil)
}

func (gpuUtilSpecDesc) getMetricDesc() *prometheus.Desc {
	return prometheus.NewDesc("container_request_gpu_utilization", "request of gpu utilization", append(defaultMetricLabels, "req_of_gpu"), nil)
}

func (gpuMemoryDesc) getDescribeDesc() *prometheus.Desc {
	return prometheus.NewDesc("container_gpu_memory_total", "gpu memory usage in MiB", []string{"gpu_memory"}, nil)
}

func (gpuMemorySpecDesc) getDescribeDesc() *prometheus.Desc {
	return prometheus.NewDesc("container_request_gpu_memory", "request of gpu memory in MiB", []string{"req_of_gpu_memory"}, nil)
}

func (gpuMemoryDesc) getMetricDesc() *prometheus.Desc {
	return prometheus.NewDesc("container_gpu_memory_total", "gpu memory usage in MiB", append(defaultMetricLabels, "gpu_memory"), nil)
}

func (gpuMemorySpecDesc) getMetricDesc() *prometheus.Desc {
	return prometheus.NewDesc("container_request_gpu_memory", "request of gpu memory in MiB", append(defaultMetricLabels, "req_of_gpu_memory"), nil)
}

// Describe implements prometheus Collector interface
func (disp *Display) Describe(ch chan<- *prometheus.Desc) {
	ch <- utilDescBuilder.getDescribeDesc()
	ch <- utilSpecDescBuilder.getDescribeDesc()
	ch <- memoryDescBuilder.getDescribeDesc()
	ch <- memorySpecDescBuilder.getDescribeDesc()
}

// Collect implements prometheus Collector interface
func (disp *Display) Collect(ch chan<- prometheus.Metric) {
	for _, pod := range watchdog.GetActivePods() {
		valueLabels := make([]string, len(defaultMetricLabels))
		valueLabels[metricPodName] = pod.Name
		valueLabels[metricNamespace] = pod.Namespace
		valueLabels[metricNodeName] = pod.Spec.NodeName

		podUsage := disp.getPodUsage(pod)
		podSpec := disp.getPodSpec(pod, podUsage)
		// Usage of container
		for contName, devicesStat := range podUsage {
			valueLabels[metricContainerName] = contName

			var totalUtils, totalMemory float32
			for _, perDeviceStat := range devicesStat.Dev {
				totalUtils += perDeviceStat.Gpu
				totalMemory += perDeviceStat.Mem

				gpuID := fmt.Sprintf("gpu%s", perDeviceStat.CardIdx)
				if perDeviceStat.Gpu >= 0 {
					ch <- prometheus.MustNewConstMetric(utilDescBuilder.getMetricDesc(),
						prometheus.GaugeValue, float64(perDeviceStat.Gpu), append(valueLabels, gpuID)...)
				}

				if perDeviceStat.Mem >= 0 {
					ch <- prometheus.MustNewConstMetric(memoryDescBuilder.getMetricDesc(),
						prometheus.GaugeValue, float64(perDeviceStat.Mem), append(valueLabels, gpuID)...)
				}
			}

			if totalUtils >= 0 {
				ch <- prometheus.MustNewConstMetric(utilDescBuilder.getMetricDesc(),
					prometheus.GaugeValue, float64(totalUtils), append(valueLabels, "total")...)
			}

			if totalMemory >= 0 {
				ch <- prometheus.MustNewConstMetric(memoryDescBuilder.getMetricDesc(),
					prometheus.GaugeValue, float64(totalMemory), append(valueLabels, "total")...)
			}
		}
		// Spec of container
		for contName, spec := range podSpec {
			valueLabels[metricContainerName] = contName

			ch <- prometheus.MustNewConstMetric(utilSpecDescBuilder.getMetricDesc(),
				prometheus.GaugeValue, float64(spec.Gpu), append(valueLabels, "total")...)
			ch <- prometheus.MustNewConstMetric(memorySpecDescBuilder.getMetricDesc(),
				prometheus.GaugeValue, float64(spec.Mem), append(valueLabels, "total")...)
		}
	}
}
