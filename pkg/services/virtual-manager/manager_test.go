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

package vitrual_manager

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	vcudaapi "tkestack.io/gpu-manager/pkg/api/runtime/vcuda"
	"tkestack.io/gpu-manager/pkg/config"
	"tkestack.io/gpu-manager/pkg/services/watchdog"
	"tkestack.io/gpu-manager/pkg/types"
	"tkestack.io/gpu-manager/pkg/utils"

	dockercontainer "github.com/docker/docker/api/types/container"
	"google.golang.org/grpc"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/pkg/kubelet/dockershim/libdocker"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

// #lizard forgives
func TestVirtualManager(t *testing.T) {
	flag.Parse()
	baseDir, _ := ioutil.TempDir("", "vm")
	virtualManager := NewVirtualManagerForTest(&config.Config{
		VirtualManagerPath: baseDir,
		VCudaRequestsQueue: make(chan *types.VCudaRequest, 10),
	})
	virtualManager.hostName = "test.com"

	procsReaderWriter := utils.NewCgroupProcs(baseDir, "")
	virtualManager.procsReader = procsReaderWriter

	defer func() {
		if t.Failed() {
			t.Logf("manager directory: %s", baseDir)
			return
		}

		os.RemoveAll(baseDir)
	}()

	fakeK8sclient := fake.NewSimpleClientset()
	// create watchdog and run
	watchdog.NewPodCacheForTest(fakeK8sclient)

	testCases := []struct {
		PodUID        string
		ContainerName string
		ContainerID   string
		Old           bool
		Recover       bool
		Pids          []int
		NodeName      string
	}{
		{
			PodUID:        "uid-0",
			ContainerName: "/k8s_container-0",
			ContainerID:   "0",
			Old:           false,
			Recover:       false,
			Pids:          []int{0},
			NodeName:      virtualManager.hostName,
		},
		{
			PodUID:        "uid-1",
			ContainerName: "/k8s_container-1",
			ContainerID:   "1",
			Old:           false,
			Recover:       true,
			Pids:          []int{1},
			NodeName:      virtualManager.hostName,
		},
		{
			PodUID:        "uid-2",
			ContainerName: "container-2",
			ContainerID:   "2",
			Old:           true,
			Recover:       false,
			Pids:          []int{2},
			NodeName:      virtualManager.hostName,
		},
		{
			PodUID:        "uid-3",
			ContainerName: "container-3",
			ContainerID:   "3",
			Old:           true,
			Recover:       true,
			Pids:          []int{3},
			NodeName:      "abc.com",
		},
	}

	fakeDockerClient := virtualManager.dockerClient.(*libdocker.FakeDockerClient)
	fakeRunningContainers := make([]*libdocker.FakeContainer, len(testCases))
	for i, cs := range testCases {
		fakeRunningContainers[i] = &libdocker.FakeContainer{
			ID:   cs.ContainerID,
			Name: cs.ContainerName,
			HostConfig: &dockercontainer.HostConfig{
				Resources: dockercontainer.Resources{
					CgroupParent: "/" + cs.PodUID,
				},
			},
		}
		if !cs.Old {
			fakeRunningContainers[i].Name = utils.MakeContainerNamePrefix(cs.ContainerName)
		}

		fakePod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cs.PodUID,
				UID:       k8stypes.UID(cs.PodUID),
				Namespace: "test",
			},
			Spec: v1.PodSpec{
				NodeName: cs.NodeName,
				Containers: []v1.Container{
					{
						Name: cs.ContainerName,
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								types.VCoreAnnotation:   resource.MustParse(fmt.Sprintf("%d", i+1)),
								types.VMemoryAnnotation: resource.MustParse(fmt.Sprintf("%d", i+1)),
							},
						},
					},
				}},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:        cs.ContainerName,
						ContainerID: cs.ContainerID,
					},
				},
			},
		}

		if _, err := fakeK8sclient.CoreV1().Pods("test").Create(fakePod); err != nil {
			t.Errorf("can't create pod %s", cs.PodUID)
		}

		dirName := ""
		if cs.Old {
			dirName = filepath.Join(baseDir, cs.PodUID, cs.ContainerName)
		} else {
			dirName = filepath.Join(baseDir, cs.PodUID)
		}

		if err := os.MkdirAll(dirName, DEFAULT_DIR_MODE); err != nil {
			t.Errorf("recover: can't mkdir for pod %s, cont: %s, %v", cs.PodUID, cs.ContainerName, err)
		}

		procsDir := filepath.Join(baseDir, cs.PodUID, cs.ContainerID)
		if err := os.MkdirAll(procsDir, DEFAULT_DIR_MODE); err != nil {
			t.Errorf("procs: can't mkdir for pod %s, cont: %s, %v", cs.PodUID, cs.ContainerName, err)
		}
		if err := procsReaderWriter.Write(cs.PodUID, cs.ContainerID, cs.Pids); err != nil {
			t.Errorf("procs: can't write pod %s, cont: %s, %v", cs.PodUID, cs.ContainerName, err)
		}
	}
	fakeDockerClient.SetFakeRunningContainers(fakeRunningContainers)

	virtualManager.Run()

	t.Logf("Test recover")
	for _, cs := range testCases {
		if cs.Recover {
			dirName := ""
			if cs.Old {
				dirName = filepath.Join(baseDir, cs.PodUID, cs.ContainerName)
			} else {
				dirName = filepath.Join(baseDir, cs.PodUID)
			}

			if _, err := os.Stat(filepath.Join(dirName, types.VDeviceSocket)); err != nil {
				if cs.NodeName == virtualManager.hostName {
					t.Errorf("can't stat %s socket file, %v", cs.PodUID, err)
				}
			}
		}
	}

	t.Logf("Test new request")
	for _, cs := range testCases {
		if !cs.Recover {
			virtualManager.cfg.VCudaRequestsQueue <- &types.VCudaRequest{
				PodUID: cs.PodUID,
			}
		}
	}

	time.After(time.Second * 5)
	for _, cs := range testCases {
		if !cs.Recover {
			if _, err := os.Stat(filepath.Join(baseDir, cs.PodUID, types.VDeviceSocket)); err != nil {
				t.Errorf("can't stat %s socket file, %v", cs.PodUID, err)
			}
		}
	}

	t.Logf("Test register")
	for _, cs := range testCases {
		if cs.NodeName != virtualManager.hostName {
			continue
		}

		socketName := ""
		request := &vcudaapi.VDeviceRequest{
			PodUid: cs.PodUID,
		}
		pidfile := ""
		cfgfile := ""

		if cs.Old {
			socketName = filepath.Join(baseDir, cs.PodUID, cs.ContainerName, types.VDeviceSocket)
			request.ContainerName = cs.ContainerName
			pidfile = filepath.Join(baseDir, cs.PodUID, cs.ContainerName, PIDS_CONFIG_NAME)
			cfgfile = filepath.Join(baseDir, cs.PodUID, cs.ContainerName, CONTROLLER_CONFIG_NAME)
		} else {
			socketName = filepath.Join(baseDir, cs.PodUID, types.VDeviceSocket)
			request.ContainerId = cs.ContainerID
			pidfile = filepath.Join(baseDir, cs.PodUID, cs.ContainerID, PIDS_CONFIG_NAME)
			cfgfile = filepath.Join(baseDir, cs.PodUID, cs.ContainerID, CONTROLLER_CONFIG_NAME)
		}

		conn, err := grpc.Dial(socketName, utils.DefaultDialOptions...)
		if err != nil {
			t.Errorf("can't dial %s", socketName)
		}

		registerClient := vcudaapi.NewVCUDAServiceClient(conn)
		_, err = registerClient.RegisterVDevice(context.Background(), request)
		if err != nil {
			t.Errorf("%s can't register to manager, %v", cs.PodUID, err)
		}

		// check pid file
		if _, err := os.Stat(pidfile); err != nil {
			t.Errorf("%s can't find pid file, %v", cs.PodUID, err)
		}

		// check config file
		if _, err := os.Stat(cfgfile); err != nil {
			t.Errorf("%s can't find config file, %v", cs.PodUID, err)
		}
	}
}
