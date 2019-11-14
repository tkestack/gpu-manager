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
	"time"

	"tkestack.io/gpu-manager/pkg/utils"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	//PodResource is a literal for pod resource
	PodResource = "pods"
)

//PodCache contains a informer of pod
type PodCache struct {
	informer cache.SharedInformer
}

var (
	podCache *PodCache
)

//NewPodCache creates a new podCache
func NewPodCache(coreClient corev1.CoreV1Interface) {
	podCache = new(PodCache)

	watcher := cache.NewListWatchFromClient(coreClient.RESTClient(), PodResource, metav1.NamespaceAll, fields.Everything())
	podCache.informer = cache.NewSharedInformer(watcher, &v1.Pod{}, time.Minute)

	podCache.informer.AddEventHandler(podCache)

	ch := make(chan struct{})
	go podCache.informer.Run(ch)

	for !podCache.informer.HasSynced() {
		time.Sleep(time.Second)
	}
	glog.V(2).Infof("Pod cache is running")
}

//NewPodCacheForTest creates a new podCache for testing
func NewPodCacheForTest(client kubernetes.Interface) {
	podCache = new(PodCache)

	informers := informers.NewSharedInformerFactory(client, 0)
	podCache.informer = informers.Core().V1().Pods().Informer()
	podCache.informer.AddEventHandler(podCache)
	ch := make(chan struct{})
	informers.Start(ch)

	for !podCache.informer.HasSynced() {
		time.Sleep(time.Second)
	}
	glog.V(2).Infof("Pod cache is running")
}

//OnAdd is a callback function for informer, do nothing for now.
func (p *PodCache) OnAdd(obj interface{}) {}

//OnUpdate is a callback function for informer, do nothing for now.
func (p *PodCache) OnUpdate(oldObj, newObj interface{}) {}

//OnDelete is a callback function for informer, do nothing for now.
func (p *PodCache) OnDelete(obj interface{}) {}

//GetActivePods get all active pods from podCache and returns them.
func GetActivePods() map[string]*v1.Pod {
	if podCache == nil {
		return nil
	}

	activePods := make(map[string]*v1.Pod)

	for _, item := range podCache.informer.GetStore().List() {
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
