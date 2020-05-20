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

package config

import (
	"time"

	"tkestack.io/gpu-manager/pkg/types"
)

// Config contains the necessary options for the plugin.
type Config struct {
	Driver                   string
	ExtraConfigPath          string
	QueryPort                int
	QueryAddr                string
	KubeConfig               string
	SamplePeriod             time.Duration
	Hostname                 string
	NodeLabels               map[string]string
	VirtualManagerPath       string
	DevicePluginPath         string
	VolumeConfigPath         string
	EnableShare              bool
	AllocationCheckPeriod    time.Duration
	CheckpointPath           string
	ContainerRuntimeEndpoint string
	CgroupDriver             string
	RequestTimeout           time.Duration

	VCudaRequestsQueue chan *types.VCudaRequest
}

//ExtraConfig contains extra options other than Config
type ExtraConfig struct {
	Devices []string `json:"devices,omitempty"`
}
