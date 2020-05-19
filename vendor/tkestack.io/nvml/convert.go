// +build !windows

/*
 * Tencent is pleased to support the open source community by making TKESStack available.
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

package nvml

/*
#cgo CFLAGS: -I/usr/local/cuda/include
#include "nvml_dl.h"
*/
import "C"

import (
	"fmt"
)

func (a RestrictedAPI) convert() C.nvmlRestrictedAPI_t {
	return C.nvmlRestrictedAPI_t(int(a))
}

func (c ClockType) convert() C.nvmlClockType_t {
	return C.nvmlClockType_t(int(c))
}

func errorString(ret C.nvmlReturn_t) error {
	if ret == C.NVML_SUCCESS {
		return nil
	}
	err := C.GoString(C.nvmlErrorString_dlib(ret))
	return fmt.Errorf("nvml: %v", err)
}

func stateBool(state C.nvmlEnableState_t) bool {
	if state == C.NVML_FEATURE_ENABLED {
		return true
	}

	return false
}

func boolState(d bool) C.nvmlEnableState_t {
	if d {
		return C.NVML_FEATURE_ENABLED
	}

	return C.NVML_FEATURE_DISABLED
}

func brandType(c C.nvmlBrandType_t) BrandType {
	return BrandType(C.brandType_to_int(c))
}

func bridgeChipType(c C.nvmlBridgeChipType_t) BridgeChipType {
	return BridgeChipType(C.bridgeChipType_to_int(c))
}

func computeModeType(c C.nvmlComputeMode_t) ComputeMode {
	return ComputeMode(C.computeMode_to_int(c))
}

func clocksThrottleReasons(c C.ulonglong) []ClocksThrottleReasons {
	d := uint64(c)
	reasons := make([]ClocksThrottleReasons, 0)

	if d&uint64(ClocksThrottleReasonAll) == uint64(ClocksThrottleReasonApplicationsClocksSetting) {
		reasons = append(reasons, ClocksThrottleReasonApplicationsClocksSetting)
	}

	if d&uint64(ClocksThrottleReasonAll) == uint64(ClocksThrottleReasonGpuIdle) {
		reasons = append(reasons, ClocksThrottleReasonGpuIdle)
	}

	if d&uint64(ClocksThrottleReasonAll) == uint64(ClocksThrottleReasonHwSlowdown) {
		reasons = append(reasons, ClocksThrottleReasonHwSlowdown)
	}

	if d&uint64(ClocksThrottleReasonAll) == uint64(ClocksThrottleReasonNone) {
		reasons = append(reasons, ClocksThrottleReasonNone)
	}

	if d&uint64(ClocksThrottleReasonAll) == uint64(ClocksThrottleReasonSwPowerCap) {
		reasons = append(reasons, ClocksThrottleReasonSwPowerCap)
	}

	if d&uint64(ClocksThrottleReasonAll) == uint64(ClocksThrottleReasonUnknown) {
		reasons = append(reasons, ClocksThrottleReasonUnknown)
	}

	return reasons
}

func (t MemoryErrorType) convert() C.nvmlMemoryErrorType_t {
	return C.nvmlMemoryErrorType_t(int(t))
}

func (t EccCounterType) convert() C.nvmlEccCounterType_t {
	return C.nvmlEccCounterType_t(int(t))
}

func gpuOperationMode(c C.nvmlGpuOperationMode_t) GpuOperationMode {
	return GpuOperationMode(C.gpuOperationMode_to_int(c))
}

func (t InforomObject) convert() C.nvmlInforomObject_t {
	return C.nvmlInforomObject_t(int(t))
}

func (t MemoryLocation) convert() C.nvmlMemoryLocation_t {
	return C.nvmlMemoryLocation_t(int(t))
}

func (t PcieUtilCounter) convert() C.nvmlPcieUtilCounter_t {
	return C.nvmlPcieUtilCounter_t(int(t))
}

func (t PageRetirementCause) convert() C.nvmlPageRetirementCause_t {
	return C.nvmlPageRetirementCause_t(int(t))
}

func (t TemperatureThresholds) convert() C.nvmlTemperatureThresholds_t {
	return C.nvmlTemperatureThresholds_t(int(t))
}

func (t GpuTopologyLevel) convert() C.nvmlGpuTopologyLevel_t {
	return C.nvmlGpuTopologyLevel_t(int(t))
}

func (t PerfPolicy) convert() C.nvmlPerfPolicyType_t {
	return C.nvmlPerfPolicyType_t(int(t))
}

func (t ComputeMode) convert() C.nvmlComputeMode_t {
	return C.nvmlComputeMode_t(int(t))
}

func (t GpuOperationMode) convert() C.nvmlGpuOperationMode_t {
	return C.nvmlGpuOperationMode_t(int(t))
}

func eventTypes(c C.ulonglong) []EventType {
	d := uint64(c)
	evtTypes := make([]EventType, 0)

	if d&uint64(EventTypeNone) == uint64(EventTypeNone) {
		evtTypes = append(evtTypes, EventTypeNone)
	}

	if d&uint64(EventTypeSingleBitEccError) == uint64(EventTypeSingleBitEccError) {
		evtTypes = append(evtTypes, EventTypeSingleBitEccError)
	}

	if d&uint64(EventTypeDoubleBitEccError) == uint64(EventTypeDoubleBitEccError) {
		evtTypes = append(evtTypes, EventTypeDoubleBitEccError)
	}

	if d&uint64(EventTypePState) == uint64(EventTypePState) {
		evtTypes = append(evtTypes, EventTypePState)
	}

	if d&uint64(EventTypeXidCriticalError) == uint64(EventTypeXidCriticalError) {
		evtTypes = append(evtTypes, EventTypeXidCriticalError)
	}

	if d&uint64(EventTypeClock) == uint64(EventTypeClock) {
		evtTypes = append(evtTypes, EventTypeClock)
	}

	return evtTypes
}
