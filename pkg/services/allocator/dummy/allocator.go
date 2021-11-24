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

package dummy

import (
	"context"
	"fmt"
	"time"

	"tkestack.io/gpu-manager/pkg/config"
	"tkestack.io/gpu-manager/pkg/device"
	"tkestack.io/gpu-manager/pkg/services/response"

	// Register test allocator controller
	_ "tkestack.io/gpu-manager/pkg/device/dummy"
	"tkestack.io/gpu-manager/pkg/services/allocator"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func init() {
	allocator.Register("dummy", NewDummyAllocator)
}

//DummyAllocator is a struct{}
type DummyAllocator struct {
}

var _ allocator.GPUTopoService = &DummyAllocator{}

//NewDummyAllocator returns a new DummyAllocator
func NewDummyAllocator(_ *config.Config, _ device.GPUTree, _ kubernetes.Interface, _ response.Manager) allocator.GPUTopoService {
	return &DummyAllocator{}
}

//Allocate returns /dev/fuse for dummy device
func (ta *DummyAllocator) Allocate(_ context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	resps := &pluginapi.AllocateResponse{}
	for range reqs.ContainerRequests {
		resps.ContainerResponses = append(resps.ContainerResponses, &pluginapi.ContainerAllocateResponse{
			Devices: []*pluginapi.DeviceSpec{
				{
					// We use /dev/fuse for dummy device
					ContainerPath: "/dev/fuse",
					HostPath:      "/dev/fuse",
					Permissions:   "mrw",
				},
			},
		})
	}

	return resps, nil
}

//ListAndWatch not implement
func (ta *DummyAllocator) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	return fmt.Errorf("not implement")
}

//ListAndWatchWithResourceName sends dummy device back to server
func (ta *DummyAllocator) ListAndWatchWithResourceName(resourceName string, e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	devs := []*pluginapi.Device{
		{
			ID:     fmt.Sprintf("dummy-%s-0", resourceName),
			Health: pluginapi.Healthy,
		},
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
func (ta *DummyAllocator) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

//PreStartContainer returns empty PreStartContainerResponse
func (ta *DummyAllocator) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}
