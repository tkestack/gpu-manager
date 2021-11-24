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

package watchdog

import (
	"os"
	"regexp"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	"tkestack.io/nvml"
)

const (
	gpuModelLabel = "gaia.tencent.com/gpu-model"
)

type labelFunc interface {
	GetLabel() string
}

type nodeLabeler struct {
	hostName    string
	client      v1core.CoreV1Interface
	labelMapper map[string]labelFunc
}

type modelFunc struct{}
type stringFunc string

var modelFn = modelFunc{}

func (m modelFunc) GetLabel() (model string) {
	if err := nvml.Init(); err != nil {
		klog.Warningf("Can't initialize nvml library, %v", err)
		return
	}

	defer nvml.Shutdown()

	// Assume all devices on this node are the same model
	dev, err := nvml.DeviceGetHandleByIndex(0)
	if err != nil {
		klog.Warningf("Can't get device 0 information, %v", err)
		return
	}

	rawName, err := dev.DeviceGetName()
	if err != nil {
		klog.Warningf("Can't get device name, %v", err)
		return
	}

	klog.V(4).Infof("GPU name: %s", rawName)

	return getTypeName(rawName)
}

func (s stringFunc) GetLabel() string {
	return string(s)
}

var modelNameSplitPattern = regexp.MustCompile("\\s+")

func getTypeName(name string) string {
	splits := modelNameSplitPattern.Split(name, -1)

	if len(splits) > 2 {
		return splits[1]
	}

	klog.V(4).Infof("GPU name splits: %v", splits)

	return ""
}

//NewNodeLabeler returns a new nodeLabeler
func NewNodeLabeler(client v1core.CoreV1Interface, hostname string, labels map[string]string) *nodeLabeler {
	if len(hostname) == 0 {
		hostname, _ = os.Hostname()
	}

	klog.V(2).Infof("Labeler for hostname %s", hostname)

	labelMapper := make(map[string]labelFunc)
	for k, v := range labels {
		if k == gpuModelLabel {
			labelMapper[k] = modelFn
		} else {
			labelMapper[k] = stringFunc(v)
		}
	}

	return &nodeLabeler{
		hostName:    hostname,
		client:      client,
		labelMapper: labelMapper,
	}
}

func (nl *nodeLabeler) Run() error {
	err := wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
		node, err := nl.client.Nodes().Get(nl.hostName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		for k, fn := range nl.labelMapper {
			l := fn.GetLabel()
			if len(l) == 0 {
				klog.Warningf("Empty label for %s", k)
				continue
			}

			klog.V(2).Infof("Label %s %s=%s", nl.hostName, k, l)
			node.Labels[k] = l
		}

		_, updateErr := nl.client.Nodes().Update(node)
		if updateErr != nil {
			if errors.IsConflict(updateErr) {
				return false, nil
			}
			return true, updateErr
		}

		return true, nil
	})

	if err != nil {
		return err
	}

	klog.V(2).Infof("Auto label is running")

	return nil
}
