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
	"flag"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

func TestNodeLabeler(t *testing.T) {
	flag.Parse()
	nodeName := "testnode"
	testKey := "testkey"
	testValue := "testvalue"
	labels := make(map[string]string)
	labels[testKey] = testValue

	// create node with fake client
	k8sclient := fake.NewSimpleClientset()
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nodeName,
			Labels: make(map[string]string),
		},
	}
	k8sclient.CoreV1().Nodes().Create(node)

	// create nodeLabeler and run
	nodeLabeler := NewNodeLabeler(k8sclient.CoreV1(), nodeName, labels)
	go nodeLabeler.Run()

	// check if nodeLabeler work well
	err := wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
		node, err := k8sclient.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if v, ok := node.Labels[testKey]; !ok || v != testValue {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		t.Fatalf("test failed: %s", err.Error())
	}
}
