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

package cache

//Info contains infomations aboud GPU
type Info struct {
	Devices []string
	Cores   int64
	Memory  int64
}

type containerToInfo map[string]*Info

// PodCache represents a list of pod to GPU mappings.
type PodCache struct {
	PodGPUMapping map[string]containerToInfo
}

//NewAllocateCache creates new PodCache
func NewAllocateCache() *PodCache {
	return &PodCache{
		PodGPUMapping: make(map[string]containerToInfo),
	}
}

//Pods returns all pods in PodCache
func (pgpu *PodCache) Pods() []string {
	ret := make([]string, 0)
	for k := range pgpu.PodGPUMapping {
		ret = append(ret, k)
	}
	return ret
}

//Insert adds GPU info of pod into PodCache if not exist
func (pgpu *PodCache) Insert(podUID, contName string, cache *Info) {
	if _, exists := pgpu.PodGPUMapping[podUID]; !exists {
		pgpu.PodGPUMapping[podUID] = make(containerToInfo)
	}
	pgpu.PodGPUMapping[podUID][contName] = cache
}

//GetCache returns GPU of pod if exist
func (pgpu *PodCache) GetCache(podUID string) map[string]*Info {
	containers, exists := pgpu.PodGPUMapping[podUID]
	if !exists {
		return nil
	}

	return containers
}

//Delete removes GPU info in PodCache
func (pgpu *PodCache) Delete(uid string) {
	delete(pgpu.PodGPUMapping, uid)
}
