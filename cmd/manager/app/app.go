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

package app

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tkestack.io/gpu-manager/cmd/manager/options"
	"tkestack.io/gpu-manager/pkg/config"
	"tkestack.io/gpu-manager/pkg/server"
	"tkestack.io/gpu-manager/pkg/types"
	"tkestack.io/gpu-manager/pkg/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

// #lizard forgives
func Run(opt *options.Options) error {
	cfg := &config.Config{
		Driver:                opt.Driver,
		QueryPort:             opt.QueryPort,
		QueryAddr:             opt.QueryAddr,
		KubeConfig:            opt.KubeConfigFile,
		Standalone:            opt.Standalone,
		SamplePeriod:          time.Duration(opt.SamplePeriod) * time.Second,
		DockerEndpoint:        opt.DockerEndpoint,
		VCudaRequestsQueue:    make(chan *types.VCudaRequest, 10),
		DevicePluginPath:      pluginapi.DevicePluginPath,
		VirtualManagerPath:    opt.VirtualManagerPath,
		VolumeConfigPath:      opt.VolumeConfigPath,
		EnableShare:           opt.EnableShare,
		AllocationCheckPeriod: time.Duration(opt.AllocationCheckPeriod) * time.Second,
		InClusterMode:         opt.InClusterMode,
		CheckpointPath:        opt.CheckpointPath,
	}

	if len(opt.HostnameOverride) > 0 {
		cfg.Hostname = opt.HostnameOverride
	}

	if len(opt.ExtraPath) > 0 {
		cfg.ExtraConfigPath = opt.ExtraPath
	}

	if len(opt.DevicePluginPath) > 0 {
		cfg.DevicePluginPath = opt.DevicePluginPath
	}

	cfg.NodeLabels = make(map[string]string)
	for _, item := range strings.Split(opt.NodeLabels, ",") {
		if len(item) > 0 {
			kvs := strings.SplitN(item, "=", 2)
			if len(kvs) == 2 {
				cfg.NodeLabels[kvs[0]] = kvs[1]
			} else {
				glog.Warningf("malformed node labels: %v", kvs)
			}
		}
	}

	srv := server.NewManager(cfg)
	go srv.Run()

	retries := 0
	for !srv.Ready() && retries < 10 {
		glog.Infof("Wait for internal server ready")
		time.Sleep(time.Second)
		retries++
	}

	if retries == 10 {
		glog.Warningf("Wait too long for server ready, restarting")
		os.Exit(1)
	}

	if err := srv.RegisterToKubelet(); err != nil {
		return err
	}

	devicePluginSocket := filepath.Join(cfg.DevicePluginPath, types.KubeletSocket)
	watcher, err := utils.NewFSWatcher(cfg.DevicePluginPath)
	if err != nil {
		log.Println("Failed to created FS watcher.")
		os.Exit(1)
	}
	defer watcher.Close()

	for {
		select {
		case event := <-watcher.Events:
			if event.Name == devicePluginSocket && event.Op&fsnotify.Create == fsnotify.Create {
				time.Sleep(time.Second)
				glog.Fatalf("inotify: %s created, restarting.", devicePluginSocket)
			}
		case err := <-watcher.Errors:
			glog.Fatalf("inotify: %s", err)
		}
	}
	return nil
}
