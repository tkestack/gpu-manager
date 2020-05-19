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

package main

import (
	goflag "flag"
	"fmt"
	"os"

	"k8s.io/klog"

	"tkestack.io/gpu-manager/cmd/manager/app"
	"tkestack.io/gpu-manager/cmd/manager/options"
	"tkestack.io/gpu-manager/pkg/flags"
	"tkestack.io/gpu-manager/pkg/logs"
	"tkestack.io/gpu-manager/pkg/version"

	"github.com/spf13/pflag"
)

func main() {
	klog.InitFlags(nil)
	opt := options.NewOptions()
	opt.AddFlags(pflag.CommandLine)

	flags.InitFlags()
	goflag.CommandLine.Parse([]string{})
	logs.InitLogs()
	defer logs.FlushLogs()

	version.PrintAndExitIfRequested()

	if err := app.Run(opt); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
