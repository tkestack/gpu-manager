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
	"net"
	"os"
	"path/filepath"
	"syscall"

	"google.golang.org/grpc"
	"k8s.io/klog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"tkestack.io/gpu-manager/pkg/types"
)

const (
	vmemorySocketName = "vmemory.sock"
)

type vmemoryResourceServer struct {
	resourceServerImpl
}

var _ pluginapi.DevicePluginServer = &vmemoryResourceServer{}
var _ ResourceServer = &vmemoryResourceServer{}

func newVmemoryServer(manager *managerImpl) ResourceServer {
	socketFile := filepath.Join(manager.config.DevicePluginPath, vmemorySocketName)
	return &vmemoryResourceServer{
		resourceServerImpl: resourceServerImpl{
			srv:        grpc.NewServer(),
			socketFile: socketFile,
			mgr:        manager,
		},
	}
}

func (vr *vmemoryResourceServer) SocketName() string {
	return vr.socketFile
}

func (vr *vmemoryResourceServer) ResourceName() string {
	return types.VMemoryAnnotation
}

func (vr *vmemoryResourceServer) Stop() {
	vr.srv.Stop()
}

func (vr *vmemoryResourceServer) Run() error {
	pluginapi.RegisterDevicePluginServer(vr.srv, vr)

	err := syscall.Unlink(vr.socketFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	l, err := net.Listen("unix", vr.socketFile)
	if err != nil {
		return err
	}

	klog.V(2).Infof("Server %s is ready at %s", types.VMemoryAnnotation, vr.socketFile)

	return vr.srv.Serve(l)
}

/** device plugin interface */
func (vr *vmemoryResourceServer) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	klog.V(2).Infof("%+v allocation request for vmemory", reqs)
	fakeData := make([]*pluginapi.ContainerAllocateResponse, 0)
	fakeData = append(fakeData, &pluginapi.ContainerAllocateResponse{})

	return &pluginapi.AllocateResponse{
		ContainerResponses: fakeData,
	}, nil
}

func (vr *vmemoryResourceServer) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	klog.V(2).Infof("ListAndWatch request for vmemory")
	return vr.mgr.ListAndWatchWithResourceName(types.VMemoryAnnotation, e, s)
}

func (vr *vmemoryResourceServer) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	klog.V(2).Infof("GetDevicePluginOptions request for vmemory")
	return &pluginapi.DevicePluginOptions{}, nil
}

func (vr *vmemoryResourceServer) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	klog.V(2).Infof("PreStartContainer request for vmemory")
	return &pluginapi.PreStartContainerResponse{}, nil
}
