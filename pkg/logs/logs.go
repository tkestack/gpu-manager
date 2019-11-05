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

package logs

import (
	"log"
	"time"

	"github.com/golang/glog"
	"google.golang.org/grpc/grpclog"
)

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (gw GlogWriter) Write(data []byte) (n int, err error) {
	glog.Info(string(data))
	return len(data), nil
}

// InitLogs initializes logs the way we want for kubernetes.
func InitLogs() {
	logger := GlogWriter{}
	log.SetOutput(logger)
	log.SetFlags(0)

	grpclog.SetLogger(logger)
	// The default glog flush interval is 30 seconds, which is frighteningly long.
	go func() {
		for range time.Tick(time.Second) {
			glog.Flush()
		}
	}()
}

//FlushLogs calls glog.Flush to flush all pending log I/O
func FlushLogs() {
	glog.Flush()
}

//Fatal wraps glog.FatalDepth
func (gw GlogWriter) Fatal(args ...interface{}) {
	glog.FatalDepth(1, args...)
}

//Fatalf wraps glog.Fatalf
func (gw GlogWriter) Fatalf(format string, args ...interface{}) {
	glog.Fatalf(format, args...)
}

//Fatalln wraps glog.Fatalln
func (gw GlogWriter) Fatalln(args ...interface{}) {
	glog.Fatalln(args...)
}

//Print wraps glog.InfoDepth
func (gw GlogWriter) Print(args ...interface{}) {
	glog.InfoDepth(1, args...)
}

//Printf wraps glog.V(2).Infof
func (gw GlogWriter) Printf(format string, args ...interface{}) {
	glog.V(2).Infof(format, args...)
}

//Println wraps glog.Info
func (gw GlogWriter) Println(args ...interface{}) {
	glog.Info(args...)
}
