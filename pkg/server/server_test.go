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

package server

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"tkestack.io/gpu-manager/cmd/manager/options"
	"tkestack.io/gpu-manager/pkg/config"
	deviceFactory "tkestack.io/gpu-manager/pkg/device"
	"tkestack.io/gpu-manager/pkg/device/nvidia"
	"tkestack.io/gpu-manager/pkg/runtime"
	allocFactory "tkestack.io/gpu-manager/pkg/services/allocator"
	"tkestack.io/gpu-manager/pkg/services/response"
	virtual_manager "tkestack.io/gpu-manager/pkg/services/virtual-manager"
	"tkestack.io/gpu-manager/pkg/services/watchdog"
	"tkestack.io/gpu-manager/pkg/types"
	"tkestack.io/gpu-manager/pkg/utils"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

type kubeletStub struct {
	sync.Mutex
	socket          string
	pluginEndpoints map[string]string
	server          *grpc.Server
}

type podRawInfo struct {
	Name       string
	UID        string
	Containers []containerRawInfo
}

type containerRawInfo struct {
	Name   string
	Cores  int
	Memory int
}

// newKubeletStub returns an initialized kubeletStub for testing purpose.
func newKubeletStub(socket string) *kubeletStub {
	return &kubeletStub{
		socket:          socket,
		pluginEndpoints: make(map[string]string),
	}
}

// Minimal implementation of deviceplugin.RegistrationServer interface
func (k *kubeletStub) Register(ctx context.Context, r *pluginapi.RegisterRequest) (*pluginapi.Empty, error) {
	k.Lock()
	defer k.Unlock()
	k.pluginEndpoints[r.ResourceName] = r.Endpoint
	return &pluginapi.Empty{}, nil
}

func (k *kubeletStub) start() error {
	os.Remove(k.socket)
	s, err := net.Listen("unix", k.socket)
	if err != nil {
		return errors.Wrap(err, "Can't listen at the socket")
	}

	k.server = grpc.NewServer()

	pluginapi.RegisterRegistrationServer(k.server, k)
	go k.server.Serve(s)

	// Wait till the grpcServer is ready to serve services.
	return utils.WaitForServer(k.socket)
}

//stop servers and clean up
func stopServer(srv *managerImpl) {
	for _, s := range srv.bundleServer {
		s.Stop()
	}
	srv.srv.Stop()
	os.RemoveAll(srv.config.VirtualManagerPath)
}

func TestServer(t *testing.T) {
	flag.Parse()
	tempDir, _ := ioutil.TempDir("", "gpu-manager")

	//init opt and cfg
	opt := options.NewOptions()
	opt.VirtualManagerPath = filepath.Clean(filepath.Join(tempDir, "vm"))
	opt.DevicePluginPath = tempDir
	opt.EnableShare = true
	opt.HostnameOverride = "testnode"
	cfg := &config.Config{
		Driver:                opt.Driver,
		QueryPort:             opt.QueryPort,
		QueryAddr:             opt.QueryAddr,
		KubeConfig:            opt.KubeConfigFile,
		SamplePeriod:          time.Duration(opt.SamplePeriod) * time.Second,
		VCudaRequestsQueue:    make(chan *types.VCudaRequest, 10),
		DevicePluginPath:      opt.DevicePluginPath,
		VirtualManagerPath:    opt.VirtualManagerPath,
		VolumeConfigPath:      opt.VolumeConfigPath,
		EnableShare:           opt.EnableShare,
		Hostname:              opt.HostnameOverride,
		AllocationCheckPeriod: 5 * time.Second,
	}

	defer func() {
		os.RemoveAll(tempDir)
	}()

	//init kubletstub
	kubeletSocket := filepath.Join(cfg.DevicePluginPath, "kubelet.sock")
	kubelet := newKubeletStub(kubeletSocket)
	err := kubelet.start()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer kubelet.server.Stop()

	// init manager
	srv, _ := NewManager(cfg).(*managerImpl)
	fakeRuntimeManager := runtime.NewContainerRuntimeManagerStub()
	srv.virtualManager = virtual_manager.NewVirtualManagerForTest(cfg, fakeRuntimeManager, response.NewFakeResponseManager())
	srv.virtualManager.Run()
	defer stopServer(srv)

	treeInitFn := deviceFactory.NewFuncForName(cfg.Driver)
	obj := treeInitFn(cfg)
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

	k8sClient := fake.NewSimpleClientset()
	watchdog.NewPodCacheForTest(k8sClient)
	initAllocator := allocFactory.NewFuncForName(cfg.Driver + "_test")
	srv.allocator = initAllocator(cfg, tree, k8sClient, response.NewFakeResponseManager())
	srv.setupGRPCService()
	srv.RegisterToKubelet()
	for _, rs := range srv.bundleServer {
		go rs.Run()
		if err := utils.WaitForServer(rs.SocketName()); err != nil {
			t.Fatalf("%s failed to start: %+v", rs.SocketName(), err)
		}
	}

	//check if bundleServers register to kublet correctly
	expectEndpoints := make(map[string]string)
	expectEndpoints[types.VCoreAnnotation] = vcoreSocketName
	expectEndpoints[types.VMemoryAnnotation] = vmemorySocketName
	if !reflect.DeepEqual(expectEndpoints, kubelet.pluginEndpoints) {
		t.Fatalf("register to kublet wrong, expect %v, got %v", expectEndpoints, kubelet.pluginEndpoints)
	}

	//check if bundleServer work correctly
	pluginSocket := filepath.Join(opt.DevicePluginPath, kubelet.pluginEndpoints[types.VCoreAnnotation])
	conn, err := grpc.Dial(pluginSocket, utils.DefaultDialOptions...)
	if err != nil {
		t.Fatalf("Failed to get connection: %+v", err)
	}
	defer conn.Close()

	//create pod with gpu resource required
	testCases := []podRawInfo{
		{
			Name: "pod-0",
			UID:  "uid-0",
			Containers: []containerRawInfo{
				{
					Name:   "container-0",
					Cores:  10,
					Memory: 1,
				},
				{
					Name:   "container-1",
					Cores:  10,
					Memory: 1,
				},
			},
		},
	}
	for _, cs := range testCases {
		containers := []corev1.Container{}
		for _, c := range cs.Containers {
			container := corev1.Container{
				Name: c.Name,
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						types.VCoreAnnotation:   resource.MustParse(fmt.Sprintf("%d", c.Cores)),
						types.VMemoryAnnotation: resource.MustParse(fmt.Sprintf("%d", c.Memory)),
					},
				},
			}
			containers = append(containers, container)
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        cs.Name,
				UID:         k8stypes.UID(cs.UID),
				Annotations: make(map[string]string),
			},
			Spec: corev1.PodSpec{
				Containers: containers,
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		}
		pod.Annotations[types.PredicateTimeAnnotation] = "1"
		pod.Annotations[types.GPUAssigned] = "false"
		for i := range pod.Spec.Containers {
			pod.Annotations[types.PredicateGPUIndexPrefix+strconv.Itoa(i)] = "0"
		}
		pod, _ = k8sClient.CoreV1().Pods("test-ns").Create(pod)

		// wait for podLister to sync
		time.Sleep(time.Second * 2)

		client := pluginapi.NewDevicePluginClient(conn)
		for _, c := range pod.Spec.Containers {
			devicesIDs := []string{}
			vcore := c.Resources.Limits[types.VCoreAnnotation]
			for i := 0; i < int(vcore.Value()); i++ {
				devicesIDs = append(devicesIDs, types.VCoreAnnotation)
			}
			_, err = client.Allocate(context.Background(), &pluginapi.AllocateRequest{
				ContainerRequests: []*pluginapi.ContainerAllocateRequest{
					{
						DevicesIDs: devicesIDs,
					},
				},
			})
			if err != nil {
				t.Errorf("Failed to allocate for container %s due to %+v", c.Name, err)
			}
		}
	}
}
