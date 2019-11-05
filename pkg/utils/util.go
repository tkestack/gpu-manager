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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	nvtree "tkestack.io/tkestack/gpu-manager/pkg/device/nvidia"
	"tkestack.io/tkestack/gpu-manager/pkg/types"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/kubelet/dockershim"
	"k8s.io/kubernetes/pkg/kubelet/dockershim/libdocker"
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

type CgroupProcsReader interface {
	Read(cgroupParent, containerID string) []int
}

type CgroupProcsWriter interface {
	Write(cgroupParent, containerID string, pids []int) error
}

type linuxCgroupProcs struct {
	base         string
	cgroupDriver string
}

func NewCgroupProcs(cgroupMountPoint, cgroupDriver string) *linuxCgroupProcs {
	return &linuxCgroupProcs{
		base:         cgroupMountPoint,
		cgroupDriver: cgroupDriver,
	}
}

func (l *linuxCgroupProcs) getProcFile(cgroupParent, containerID string) string {
	if l.cgroupDriver == "systemd" {
		base := ""
		qos := ""

		splits := strings.SplitN(cgroupParent, "-", 3)
		if len(splits) >= 2 {
			base = splits[0]
		}

		if len(splits) == 3 {
			qos = splits[1]
		}

		return filepath.Clean(filepath.Join(l.base, base+".slice", base+"-"+qos+".slice", cgroupParent, "docker-"+containerID+".scope", types.CGROUP_PROCS))
	}

	return filepath.Clean(filepath.Join(l.base, cgroupParent, containerID, types.CGROUP_PROCS))
}

//Read pids of container from cgroup.procs file
func (l *linuxCgroupProcs) Read(cgroupParent, containerID string) []int {
	procsFile := l.getProcFile(cgroupParent, containerID)

	f, err := os.Open(procsFile)
	if err != nil {
		glog.Errorf("can't read %s, %v", procsFile, err)
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	pids := make([]int, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if pid, err := strconv.Atoi(line); err == nil {
			pids = append(pids, pid)
		}
	}

	glog.V(4).Infof("Read from %s, pids: %v", procsFile, pids)
	return pids
}

func (l *linuxCgroupProcs) Write(cgroupParent, containerID string, pids []int) error {
	procsFile := l.getProcFile(cgroupParent, containerID)

	glog.Infof("Try to write pid file at %s", procsFile)
	f, err := os.OpenFile(procsFile, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, pid := range pids {
		_, err := io.WriteString(f, fmt.Sprintf("%d\n", pid))
		if err != nil {
			return err
		}
	}

	glog.V(4).Infof("Write down %s, pids: %v", procsFile, pids)
	return nil
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

//TruncateContainerName truncate container names to name[:TruncateLen]
//if len(name) > TruncateLen
func TruncateContainerName(name string) string {
	if len(name) > TruncateLen {
		newName := name[:TruncateLen]
		glog.V(2).Infof("truncate container name from %s to %s", name, newName)
		return newName
	}

	return name
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
	glog.V(4).Infof("Try v2 checkpoint data format")
	err = json.Unmarshal(data, cpV2Data)
	if err != nil {
		return nil, err
	}

	if cpV2Data.Data != nil {
		return cpV2Data.Data, nil
	}

	glog.V(4).Infof("Try v1 checkpoint data format")
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
	glog.V(4).Infof("Determine if the pod %s needs GPU resource", pod.Name)

	vcore := GetGPUResourceOfPod(pod, types.VCoreAnnotation)
	vmemory := GetGPUResourceOfPod(pod, types.VMemoryAnnotation)

	// Check if pod request for GPU resource
	if vcore <= 0 || (vcore < nvtree.HundredCore && vmemory <= 0) {
		glog.V(4).Infof("Pod %s in namespace %s does not Request for GPU resource",
			pod.Name,
			pod.Namespace)
		return false
	}

	return true
}

func IsGPURequiredContainer(c *v1.Container) bool {
	glog.V(4).Infof("Determine if the container %s needs GPU resource", c.Name)

	vcore := GetGPUResourceOfContainer(c, types.VCoreAnnotation)
	vmemory := GetGPUResourceOfContainer(c, types.VMemoryAnnotation)

	// Check if container request for GPU resource
	if vcore <= 0 || (vcore < nvtree.HundredCore && vmemory <= 0) {
		glog.V(4).Infof("Container %s does not Request for GPU resource", c.Name)
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
	glog.V(4).Infof("Determine if the pod %s needs GPU resource", pod.Name)
	var ok bool

	// Check if pod request for GPU resource
	if GetGPUResourceOfPod(pod, types.VCoreAnnotation) <= 0 || GetGPUResourceOfPod(pod, types.VMemoryAnnotation) <= 0 {
		glog.V(4).Infof("Pod %s in namespace %s does not Request for GPU resource",
			pod.Name,
			pod.Namespace)
		return predicated
	}

	// Check if pod already has predicate time
	if _, ok = pod.ObjectMeta.Annotations[types.PredicateTimeAnnotation]; !ok {
		glog.V(4).Infof("No predicate time for pod %s in namespace %s",
			pod.Name,
			pod.Namespace)
		return predicated
	}

	// Check if pod has already been assigned
	if assigned, ok := pod.ObjectMeta.Annotations[types.GPUAssigned]; !ok {
		glog.V(4).Infof("No assigned flag for pod %s in namespace %s",
			pod.Name,
			pod.Namespace)
		return predicated
	} else if assigned == "true" {
		glog.V(4).Infof("pod %s in namespace %s has already been assigned",
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
		glog.V(4).Infof("No assigned flag for pod %s in namespace %s",
			pod.Name,
			pod.Namespace)
		return false
	} else if assigned == "false" {
		glog.V(4).Infof("pod %s in namespace %s has not been assigned",
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
			glog.Warningf("Failed to parse predicate Timestamp %s due to %v", predicateTimeStr, err)
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

func CreateDockerClient(endpoint string) libdocker.Interface {
	runtimeRequestTimeout := metav1.Duration{Duration: 2 * time.Minute}
	imagePullProgressDeadline := metav1.Duration{Duration: 1 * time.Minute}
	dockerClientConfig := &dockershim.ClientConfig{
		DockerEndpoint:            endpoint,
		RuntimeRequestTimeout:     runtimeRequestTimeout.Duration,
		ImagePullProgressDeadline: imagePullProgressDeadline.Duration,
	}

	return dockershim.NewDockerClientFromConfig(dockerClientConfig)
}
