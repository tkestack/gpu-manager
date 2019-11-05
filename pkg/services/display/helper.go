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

package display

type containerToCgroup map[string]string

// podGPUs represents a list of pod to GPU mappings.
type podGPUs struct {
	podGPUMapping map[string]containerToCgroup
}

func newPodGPUs() *podGPUs {
	return &podGPUs{
		podGPUMapping: make(map[string]containerToCgroup),
	}
}

func (pgpu *podGPUs) pods() []string {
	ret := make([]string, 0)
	for k := range pgpu.podGPUMapping {
		ret = append(ret, k)
	}
	return ret
}

func (pgpu *podGPUs) insert(podUID, contName string, cgroup string) {
	if _, exists := pgpu.podGPUMapping[podUID]; !exists {
		pgpu.podGPUMapping[podUID] = make(containerToCgroup)
	}
	pgpu.podGPUMapping[podUID][contName] = cgroup
}

func (pgpu *podGPUs) getCgroup(podUID, contName string) string {
	containers, exists := pgpu.podGPUMapping[podUID]
	if !exists {
		return ""
	}
	cgroup, exists := containers[contName]
	if !exists {
		return ""
	}
	return cgroup
}

func (pgpu *podGPUs) delete(uid string) []string {
	var cgroups []string

	for _, cont := range pgpu.podGPUMapping[uid] {
		cgroups = append(cgroups, cont)
	}

	delete(pgpu.podGPUMapping, uid)

	return cgroups
}
