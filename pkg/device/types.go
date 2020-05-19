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

package device

import (
	"tkestack.io/gpu-manager/pkg/config"

	"k8s.io/klog"
)

//GPUTree is an interface for GPU tree structure
type GPUTree interface {
	Init(input string)
	Update()
}

//NewFunc is a function to create GPUTree
type NewFunc func(cfg *config.Config) GPUTree

var (
	factory = make(map[string]NewFunc)
)

//Register NewFunc with name, which can be get
//by calling NewFuncForName() later.
func Register(name string, item NewFunc) {
	if _, ok := factory[name]; ok {
		return
	}

	klog.V(2).Infof("Register NewFunc with name %s", name)

	factory[name] = item
}

//NewFuncForName tries to find functions with specific name
//from factory, return nil if not found.
func NewFuncForName(name string) NewFunc {
	if item, ok := factory[name]; ok {
		return item
	}

	klog.V(2).Infof("Can not find NewFunc with name %s", name)

	return nil
}
