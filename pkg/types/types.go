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

package types

import (
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	VDeviceAnnotation       = "tencent.com/vcuda-device"
	VCoreAnnotation         = "tencent.com/vcuda-core"
	VCoreLimitAnnotation    = "tencent.com/vcuda-core-limit"
	VMemoryAnnotation       = "tencent.com/vcuda-memory"
	PredicateTimeAnnotation = "tencent.com/predicate-time"
	PredicateGPUIndexPrefix = "tencent.com/predicate-gpu-idx-"
	GPUAssigned             = "tencent.com/gpu-assigned"
	ClusterNameAnnotation   = "clusterName"

	VCUDA_MOUNTPOINT = "/etc/vcuda"

	/** 256MB */
	MemoryBlockSize = 268435456

	KubeletSocket                 = "kubelet.sock"
	VDeviceSocket                 = "vcuda.sock"
	CheckPointFileName            = "kubelet_internal_checkpoint"
	PreStartContainerCheckErrMsg  = "PreStartContainer check failed"
	PreStartContainerCheckErrType = "PreStartContainerCheckErr"
	UnexpectedAdmissionErrType    = "UnexpectedAdmissionError"
)

const (
	NvidiaCtlDevice    = "/dev/nvidiactl"
	NvidiaUVMDevice    = "/dev/nvidia-uvm"
	NvidiaFullpathRE   = `^/dev/nvidia([0-9]*)$`
	NvidiaDevicePrefix = "/dev/nvidia"
)

const (
	ManagerSocket = "/var/run/gpu-manager.sock"
)

const (
	CGROUP_BASE  = "/sys/fs/cgroup/memory"
	CGROUP_PROCS = "cgroup.procs"
)

type VCudaRequest struct {
	PodUID           string
	AllocateResponse *pluginapi.ContainerAllocateResponse
	ContainerName    string
	//Deprecated
	Cores int64
	//Deprecated
	Memory int64
	Done   chan error
}

type DevicesPerNUMA map[int64][]string

type PodDevicesEntry struct {
	PodUID        string
	ContainerName string
	ResourceName  string
	DeviceIDs     []string
	AllocResp     []byte
}

type PodDevicesEntryNUMA struct {
	PodUID        string
	ContainerName string
	ResourceName  string
	DeviceIDs     DevicesPerNUMA
	AllocResp     []byte
}

type CheckpointNUMA struct {
	PodDeviceEntries  []PodDevicesEntryNUMA
	RegisteredDevices map[string][]string
}

type Checkpoint struct {
	PodDeviceEntries  []PodDevicesEntry
	RegisteredDevices map[string][]string
}

type CheckpointDataNUMA struct {
	Data *CheckpointNUMA `json:"Data"`
}

type CheckpointData struct {
	Data *Checkpoint `json:"Data"`
}

var (
	DriverVersionMajor      int
	DriverVersionMinor      int
	DriverLibraryPath       string
	DriverOriginLibraryPath string
)

const (
	ContainerNameLabelKey = "io.kubernetes.container.name"
	PodNamespaceLabelKey  = "io.kubernetes.pod.namespace"
	PodNameLabelKey       = "io.kubernetes.pod.name"
	PodUIDLabelKey        = "io.kubernetes.pod.uid"
	PodCgroupNamePrefix   = "pod"
)
