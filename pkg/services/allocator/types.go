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

package allocator

import (
	"tkestack.io/gpu-manager/pkg/config"
	"tkestack.io/gpu-manager/pkg/device"
	"tkestack.io/gpu-manager/pkg/services/response"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

//GPUTopoService is server api for GPU topology service
type GPUTopoService interface {
	pluginapi.DevicePluginServer
	ListAndWatchWithResourceName(string, *pluginapi.Empty, pluginapi.DevicePlugin_ListAndWatchServer) error
}

//NewFunc represents function for creating new GPUTopoService
type NewFunc func(cfg *config.Config,
	tree device.GPUTree,
	k8sClient kubernetes.Interface,
	responseManager response.Manager) GPUTopoService

var (
	factory = make(map[string]NewFunc)
)

//Register stores NewFunc in factory
func Register(name string, item NewFunc) {
	if _, ok := factory[name]; ok {
		return
	}

	klog.V(2).Infof("Register NewFunc with name %s", name)

	factory[name] = item
}

//NewFuncForName tries to find NewFunc by name, return nil if not found
func NewFuncForName(name string) NewFunc {
	if item, ok := factory[name]; ok {
		return item
	}

	klog.V(2).Infof("Can not find NewFunc with name %s", name)

	return nil
}
