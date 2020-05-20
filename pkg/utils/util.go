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

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	nvtree "tkestack.io/gpu-manager/pkg/device/nvidia"
	"tkestack.io/gpu-manager/pkg/types"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

//constants used in this package
const (
	TruncateLen = 31
	kubePrefix  = "k8s"
)

var (
	//DefaultDialOptions contains default dial options used in grpc dial
	DefaultDialOptions = []grpc.DialOption{grpc.WithInsecure(), grpc.WithDialer(UnixDial), grpc.WithBlock()}
)

//UnixDial dials to a unix socket using net.DialTimeout
func UnixDial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", addr, timeout)
}

//IsValidGPUPath checks if path is valid Nvidia GPU device path
func IsValidGPUPath(path string) bool {
	return regexp.MustCompile(types.NvidiaFullpathRE).MatchString(path)
}

//GetGPUMinorID returns id in Nvidia GPU device path
func GetGPUMinorID(path string) (int, error) {
	str := regexp.MustCompile(types.NvidiaFullpathRE).FindStringSubmatch(path)

	if len(str) != 2 {
		return -1, fmt.Errorf("not match pattern %s", types.NvidiaFullpathRE)
	}

	id, _ := strconv.ParseInt(str[1], 10, 32)

	return int(id), nil
}

//GetGPUData get cores, memory and device names from annotations
func GetGPUData(annotations map[string]string) (gpuUtil int64, gpuMemory int64, deviceNames []string) {
	for k, v := range annotations {
		switch {
		case strings.HasSuffix(k, types.VCoreAnnotation):
			gpuUtil, _ = strconv.ParseInt(v, 10, 64)
		case strings.HasSuffix(k, types.VMemoryAnnotation):
			gpuMemory, _ = strconv.ParseInt(v, 10, 64)
		case strings.HasSuffix(k, types.VDeviceAnnotation):
			deviceNames = strings.Split(annotations[k], ",")
		}
	}

	return gpuUtil, gpuMemory, deviceNames
}

//NewFSWatcher returns a file watcher created by fsnotify.NewWatcher
func NewFSWatcher(files ...string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		err = watcher.Add(f)
		if err != nil {
			watcher.Close()
			return nil, err
		}
	}

	return watcher, nil
}

// WaitForServer checks if grpc server is alive
// by making grpc blocking connection to the server socket
func WaitForServer(socket string) error {
	conn, err := grpc.DialContext(context.Background(), socket, DefaultDialOptions...)
	if err == nil {
		conn.Close()
		return nil
	}
	return errors.Wrapf(err, "Failed dial context at %s", socket)
}

func GetCheckpointData(devicePluginPath string) (*types.Checkpoint, error) {
	cpFile := filepath.Join(devicePluginPath, types.CheckPointFileName)
	cpV2Data := &types.CheckpointData{}
	data, err := ioutil.ReadFile(cpFile)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Try v2 checkpoint data format")
	err = json.Unmarshal(data, cpV2Data)
	if err != nil {
		return nil, err
	}

	if cpV2Data.Data != nil {
		return cpV2Data.Data, nil
	}

	klog.V(4).Infof("Try v1 checkpoint data format")
	cpV1Data := &types.Checkpoint{}
	err = json.Unmarshal(data, cpV1Data)
	if err != nil {
		return nil, err
	}

	return cpV1Data, nil
}

func IsStringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func ShouldRetry(err error) bool {
	return apierr.IsConflict(err) || apierr.IsServerTimeout(err)
}

func MakeContainerNamePrefix(containerName string) string {
	return fmt.Sprintf("/%s_%s_", kubePrefix, containerName)
}

func IsGPURequiredPod(pod *v1.Pod) bool {
	vcore := GetGPUResourceOfPod(pod, types.VCoreAnnotation)
	vmemory := GetGPUResourceOfPod(pod, types.VMemoryAnnotation)

	// Check if pod request for GPU resource
	if vcore <= 0 || (vcore < nvtree.HundredCore && vmemory <= 0) {
		klog.V(4).Infof("Pod %s in namespace %s does not Request for GPU resource",
			pod.Name,
			pod.Namespace)
		return false
	}

	return true
}

func IsGPURequiredContainer(c *v1.Container) bool {
	klog.V(4).Infof("Determine if the container %s needs GPU resource", c.Name)

	vcore := GetGPUResourceOfContainer(c, types.VCoreAnnotation)
	vmemory := GetGPUResourceOfContainer(c, types.VMemoryAnnotation)

	// Check if container request for GPU resource
	if vcore <= 0 || (vcore < nvtree.HundredCore && vmemory <= 0) {
		klog.V(4).Infof("Container %s does not Request for GPU resource", c.Name)
		return false
	}

	return true
}

func GetGPUResourceOfPod(pod *v1.Pod, resourceName v1.ResourceName) uint {
	var total uint
	containers := pod.Spec.Containers
	for _, container := range containers {
		if val, ok := container.Resources.Limits[resourceName]; ok {
			total += uint(val.Value())
		}
	}
	return total
}

func ShouldDelete(pod *v1.Pod) bool {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil &&
			strings.Contains(status.State.Waiting.Message, types.PreStartContainerCheckErrMsg) {
			return true
		}
	}
	if pod.Status.Reason == types.UnexpectedAdmissionErrType {
		return true
	}
	return false
}

func IsGPUPredicatedPod(pod *v1.Pod) (predicated bool) {
	klog.V(4).Infof("Determine if the pod %s needs GPU resource", pod.Name)
	var ok bool

	// Check if pod request for GPU resource
	if GetGPUResourceOfPod(pod, types.VCoreAnnotation) <= 0 || GetGPUResourceOfPod(pod, types.VMemoryAnnotation) <= 0 {
		klog.V(4).Infof("Pod %s in namespace %s does not Request for GPU resource",
			pod.Name,
			pod.Namespace)
		return predicated
	}

	// Check if pod already has predicate time
	if _, ok = pod.ObjectMeta.Annotations[types.PredicateTimeAnnotation]; !ok {
		klog.V(4).Infof("No predicate time for pod %s in namespace %s",
			pod.Name,
			pod.Namespace)
		return predicated
	}

	// Check if pod has already been assigned
	if assigned, ok := pod.ObjectMeta.Annotations[types.GPUAssigned]; !ok {
		klog.V(4).Infof("No assigned flag for pod %s in namespace %s",
			pod.Name,
			pod.Namespace)
		return predicated
	} else if assigned == "true" {
		klog.V(4).Infof("pod %s in namespace %s has already been assigned",
			pod.Name,
			pod.Namespace)
		return predicated
	}
	predicated = true
	return predicated
}

// Check if pod has already been assigned
func IsGPUAssignedPod(pod *v1.Pod) bool {
	if assigned, ok := pod.ObjectMeta.Annotations[types.GPUAssigned]; !ok {
		klog.V(4).Infof("No assigned flag for pod %s in namespace %s",
			pod.Name,
			pod.Namespace)
		return false
	} else if assigned == "false" {
		klog.V(4).Infof("pod %s in namespace %s has not been assigned",
			pod.Name,
			pod.Namespace)
		return false
	}

	return true
}

func GetPredicateTimeOfPod(pod *v1.Pod) (predicateTime uint64) {
	if predicateTimeStr, ok := pod.ObjectMeta.Annotations[types.PredicateTimeAnnotation]; ok {
		u64, err := strconv.ParseUint(predicateTimeStr, 10, 64)
		if err != nil {
			klog.Warningf("Failed to parse predicate Timestamp %s due to %v", predicateTimeStr, err)
		} else {
			predicateTime = u64
		}
	} else {
		// If predicate time not found, use createionTimestamp instead
		predicateTime = uint64(pod.ObjectMeta.CreationTimestamp.UnixNano())
	}

	return predicateTime
}

func GetGPUResourceOfContainer(container *v1.Container, resourceName v1.ResourceName) uint {
	var count uint
	if val, ok := container.Resources.Limits[resourceName]; ok {
		count = uint(val.Value())
	}
	return count
}

func GetContainerIndexByName(pod *v1.Pod, containerName string) (int, error) {
	containerIndex := -1
	for i, c := range pod.Spec.Containers {
		if c.Name == containerName {
			containerIndex = i
			break
		}
	}

	if containerIndex == -1 {
		return containerIndex, fmt.Errorf("failed to get index of container %s in pod %s", containerName, pod.UID)
	}
	return containerIndex, nil
}
