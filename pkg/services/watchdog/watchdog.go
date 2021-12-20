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
	"fmt"
	"time"

	"tkestack.io/gpu-manager/pkg/utils"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	informerCore "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	podHostField = "spec.nodeName"
)

//PodCache contains a podInformer of pod
type PodCache struct {
	podInformer informerCore.PodInformer
}

var (
	podCache *PodCache
)

//NewPodCache creates a new podCache
func NewPodCache(client kubernetes.Interface, hostName string) {
	podCache = new(PodCache)

	factory := informers.NewSharedInformerFactoryWithOptions(client, time.Minute,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector(podHostField, hostName).String()
		}))
	podCache.podInformer = factory.Core().V1().Pods()

	ch := make(chan struct{})
	go podCache.podInformer.Informer().Run(ch)

	for !podCache.podInformer.Informer().HasSynced() {
		time.Sleep(time.Second)
	}
	klog.V(2).Infof("Pod cache is running")
}

//NewPodCacheForTest creates a new podCache for testing
func NewPodCacheForTest(client kubernetes.Interface) {
	podCache = new(PodCache)

	informers := informers.NewSharedInformerFactory(client, 0)
	podCache.podInformer = informers.Core().V1().Pods()
	podCache.podInformer.Informer().AddEventHandler(podCache)
	ch := make(chan struct{})
	informers.Start(ch)

	for !podCache.podInformer.Informer().HasSynced() {
		time.Sleep(time.Second)
	}
	klog.V(2).Infof("Pod cache is running")
}

//OnAdd is a callback function for podInformer, do nothing for now.
func (p *PodCache) OnAdd(obj interface{}) {}

//OnUpdate is a callback function for podInformer, do nothing for now.
func (p *PodCache) OnUpdate(oldObj, newObj interface{}) {}

//OnDelete is a callback function for podInformer, do nothing for now.
func (p *PodCache) OnDelete(obj interface{}) {}

//GetActivePods get all active pods from podCache and returns them.
func GetActivePods() map[string]*v1.Pod {
	if podCache == nil {
		return nil
	}

	activePods := make(map[string]*v1.Pod)

	for _, item := range podCache.podInformer.Informer().GetStore().List() {
		pod, ok := item.(*v1.Pod)
		if !ok {
			continue
		}

		if podIsTerminated(pod) {
			continue
		}

		if !utils.IsGPURequiredPod(pod) {
			continue
		}

		activePods[string(pod.UID)] = pod
	}

	return activePods
}

func GetPod(namespace, name string) (*v1.Pod, error) {
	pod, err := podCache.podInformer.Lister().Pods(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	if podIsTerminated(pod) {
		return nil, fmt.Errorf("terminated pod")
	}

	if !utils.IsGPURequiredPod(pod) {
		return nil, fmt.Errorf("no gpu pod")
	}

	return pod, nil
}

func podIsTerminated(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded || (pod.DeletionTimestamp != nil && notRunning(pod.Status.ContainerStatuses))
}

// notRunning returns true if every status is terminated or waiting, or the status list
// is empty.
func notRunning(statuses []v1.ContainerStatus) bool {
	for _, status := range statuses {
		if status.State.Terminated == nil && status.State.Waiting == nil {
			return false
		}
	}
	return true
}
