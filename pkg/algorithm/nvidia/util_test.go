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

package nvidia

import (
	"tkestack.io/gpu-manager/pkg/device/nvidia"
)

func examining(expect []string, nodes []*nvidia.NvidiaNode) (pass bool, want string, actual string) {
	if len(expect) != len(nodes) {
		return false, "", ""
	}

	for i, n := range nodes {
		if expect[i] != n.MinorName() {
			return false, expect[i], n.MinorName()
		}
	}

	return true, "", ""
}
