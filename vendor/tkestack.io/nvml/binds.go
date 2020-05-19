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

// #cgo CFLAGS: -I. -I/usr/local/cuda/include
// #include "nvml_dl.h"
import "C"

import (
	"errors"
	"runtime"
	"time"
)

func Init() error {
	r := C.nvmlInit_dlib()
	if r == C.NVML_ERROR_LIBRARY_NOT_FOUND {
		return errors.New("could not load NVML library")
	}
	return errorString(r)
}

func Shutdown() error {
	return errorString(C.nvmlShutdown_dlib())
}

func DeviceGetCount() (uint, error) {
	var n C.uint

	r := C.nvmlDeviceGetCount_dlib(&n)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(n), nil
}

func SystemGetDriverVersion() (string, error) {
	var driver [szDriver]C.char

	r := C.nvmlSystemGetDriverVersion_dlib(&driver[0], szDriver)

	if r != OP_SUCCESS {
		return "", errorString(r)
	}

	return C.GoString(&driver[0]), nil
}

func SystemGetNVMLVersion() (string, error) {
	var driver [szDriver]C.char

	r := C.nvmlSystemGetNVMLVersion_dlib(&driver[0], szDriver)

	if r != OP_SUCCESS {
		return "", errorString(r)
	}

	return C.GoString(&driver[0]), nil
}

func SystemGetProcessName(pid uint) (string, error) {
	var proc [szProcName]C.char

	r := C.nvmlSystemGetProcessName_dlib(C.uint(pid), &proc[0], szProcName)

	if r != OP_SUCCESS {
		return "", errorString(r)
	}

	return C.GoString(&proc[0]), nil
}

func (h handle) DeviceClearCpuAffinity() error {
	r := C.nvmlDeviceClearCpuAffinity_dlib(h.dev)

	return errorString(r)
}

func (h handle) DeviceGetAPIRestriction(apiType RestrictedAPI) (bool, error) {
	var state C.nvmlEnableState_t

	r := C.nvmlDeviceGetAPIRestriction_dlib(h.dev, apiType.convert(), &state)

	if r != OP_SUCCESS {
		return false, errorString(r)
	}

	return stateBool(state), nil
}

func (h handle) DeviceGetApplicationsClock(clockType ClockType) (uint, error) {
	var clockMHz C.uint

	r := C.nvmlDeviceGetApplicationsClock_dlib(h.dev, clockType.convert(), &clockMHz)
	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(clockMHz), nil
}

func (h handle) DeviceGetAutoBoostedClocksEnabled() (curState bool, defaultState bool, err error) {
	var isEnabled, defaultEnabled C.nvmlEnableState_t

	r := C.nvmlDeviceGetAutoBoostedClocksEnabled_dlib(h.dev, &isEnabled, &defaultEnabled)

	if r != OP_SUCCESS {
		return false, false, errorString(r)
	}

	return stateBool(isEnabled), stateBool(defaultEnabled), nil
}

func (h handle) DeviceGetBAR1MemoryInfo() (free uint64, used uint64, total uint64, err error) {
	var bar1 C.nvmlBAR1Memory_t

	r := C.nvmlDeviceGetBAR1MemoryInfo_dlib(h.dev, &bar1)

	if r != OP_SUCCESS {
		return 0, 0, 0, errorString(r)
	}

	return uint64(bar1.bar1Free), uint64(bar1.bar1Used), uint64(bar1.bar1Total), nil
}

func (h handle) DeviceGetBoardId() (uint, error) {
	var boardID C.uint

	r := C.nvmlDeviceGetBoardId_dlib(h.dev, &boardID)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(boardID), nil
}

func (h handle) DeviceGetBrand() (BrandType, error) {
	var brand C.nvmlBrandType_t

	r := C.nvmlDeviceGetBrand_dlib(h.dev, &brand)

	if r != OP_SUCCESS {
		return BRAND_COUNT, errorString(r)
	}

	return brandType(brand), nil
}

func (h handle) DeviceGetBridgeChipInfo() ([]*BridgeChipInfo, error) {
	var bridgeHierarchy C.nvmlBridgeChipHierarchy_t

	r := C.nvmlDeviceGetBridgeChipInfo_dlib(h.dev, &bridgeHierarchy)

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	var bridgeInfo []*BridgeChipInfo

	n := int(bridgeHierarchy.bridgeCount)
	for i := 0; i < n; i++ {
		bridgeInfo = append(bridgeInfo, &BridgeChipInfo{
			FwVersion: uint(bridgeHierarchy.bridgeChipInfo[i].fwVersion),
			Type:      bridgeChipType(bridgeHierarchy.bridgeChipInfo[i]._type),
		})
	}

	return bridgeInfo, nil
}

func (h handle) DeviceGetClockInfo(clockType ClockType) (uint, error) {
	var clockMHz C.uint

	r := C.nvmlDeviceGetClockInfo_dlib(h.dev, clockType.convert(), &clockMHz)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(clockMHz), nil
}

func (h handle) DeviceGetComputeMode() (ComputeMode, error) {
	var mode C.nvmlComputeMode_t

	r := C.nvmlDeviceGetComputeMode_dlib(h.dev, &mode)

	if r != OP_SUCCESS {
		return COMPUTEMODE_COUNT, errorString(r)
	}

	return computeModeType(mode), nil
}

func (h handle) DeviceGetComputeRunningProcesses(size int) ([]*ProcessInfo, error) {
	var procs = make([]C.nvmlProcessInfo_t, size)
	var count = C.uint(size)

	r := C.nvmlDeviceGetComputeRunningProcesses_dlib(h.dev, &count, &procs[0])
	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	n := int(count)
	info := make([]*ProcessInfo, n)
	for i := 0; i < n; i++ {
		info[i] = &ProcessInfo{
			Pid:           uint(procs[i].pid),
			UsedGPUMemory: uint64(procs[i].usedGpuMemory),
		}
	}

	return info, nil
}

func (h handle) DeviceGetCpuAffinity(size uint) ([]uint, error) {
	var d = make([]C.ulong, size)

	r := C.nvmlDeviceGetCpuAffinity_dlib(h.dev, C.uint(size), &d[0])

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	var contains = make([]uint, 0)

	if runtime.GOARCH != "amd64" {
		wordLength := uint32(32)
		cpusets := make([]uint32, size)

		for i := range d {
			cpusets[i] = uint32(d[i])
		}

		for i := uint32(0); i < uint32(size); i++ {
			for j := uint32(0); j < wordLength; j++ {
				mask := uint32(1) << j
				if (mask & cpusets[i]) > 0 {
					contains = append(contains, uint(i*wordLength+j))
				}
			}
		}

		return contains, nil
	}

	wordLength := uint64(64)
	cpusets := make([]uint64, size)

	for i := range d {
		cpusets[i] = uint64(d[i])
	}

	for i := uint64(0); i < uint64(size); i++ {
		for j := uint64(0); j < wordLength; j++ {
			mask := uint64(1) << j
			if (mask & cpusets[i]) > 0 {
				contains = append(contains, uint(i*wordLength+j))
			}
		}
	}

	return contains, nil
}

func (h handle) DeviceGetCurrPcieLinkGeneration() (uint, error) {
	var linkGen C.uint

	r := C.nvmlDeviceGetCurrPcieLinkGeneration_dlib(h.dev, &linkGen)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(linkGen), nil
}

func (h handle) DeviceGetCurrPcieLinkWidth() (uint, error) {
	var width C.uint

	r := C.nvmlDeviceGetCurrPcieLinkWidth_dlib(h.dev, &width)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(width), nil
}

func (h handle) DeviceGetCurrentClocksThrottleReasons() ([]ClocksThrottleReasons, error) {
	var reason C.ulonglong

	r := C.nvmlDeviceGetCurrentClocksThrottleReasons_dlib(h.dev, &reason)

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	return clocksThrottleReasons(reason), nil
}

func (h handle) DeviceGetDecoderUtilization() (util uint, samplePeriod uint, err error) {
	var utilization, samplingPeriodUs C.uint

	r := C.nvmlDeviceGetDecoderUtilization_dlib(h.dev, &utilization, &samplingPeriodUs)

	if r != OP_SUCCESS {
		return 0, 0, errorString(r)
	}

	return uint(utilization), uint(samplingPeriodUs), nil
}

func (h handle) DeviceGetDefaultApplicationsClock(clockType ClockType) (uint, error) {
	var clockMHz C.uint

	r := C.nvmlDeviceGetDefaultApplicationsClock_dlib(h.dev, clockType.convert(), &clockMHz)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(clockMHz), nil
}

func (h handle) DeviceGetDetailedEccErrors(mt MemoryErrorType, ec EccCounterType) (*EccErrorCounts, error) {
	var count C.nvmlEccErrorCounts_t

	r := C.nvmlDeviceGetDetailedEccErrors_dlib(h.dev, mt.convert(), ec.convert(), &count)

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	return &EccErrorCounts{
		DeviceMemory: uint64(count.deviceMemory),
		L1Cache:      uint64(count.l1Cache),
		L2Cache:      uint64(count.l2Cache),
		RegisterFile: uint64(count.registerFile),
	}, nil
}

func (h handle) DeviceGetDisplayActive() (bool, error) {
	var state C.nvmlEnableState_t

	r := C.nvmlDeviceGetDisplayActive_dlib(h.dev, &state)

	if r != OP_SUCCESS {
		return false, errorString(r)
	}

	return stateBool(state), nil
}

func (h handle) DeviceGetDisplayMode() (bool, error) {
	var state C.nvmlEnableState_t

	r := C.nvmlDeviceGetDisplayMode_dlib(h.dev, &state)

	if r != OP_SUCCESS {
		return false, errorString(r)
	}

	return stateBool(state), nil
}

func (h handle) DeviceGetEccMode() (curMode bool, pendingMode bool, err error) {
	var current, pending C.nvmlEnableState_t

	r := C.nvmlDeviceGetEccMode_dlib(h.dev, &current, &pending)

	if r != OP_SUCCESS {
		return false, false, errorString(r)
	}

	return stateBool(current), stateBool(pending), nil
}

func (h handle) DeviceGetEncoderUtilization() (util uint, samplePeriod uint, err error) {
	var utilization, samplingPeriodUs C.uint

	r := C.nvmlDeviceGetEncoderUtilization_dlib(h.dev, &utilization, &samplingPeriodUs)

	if r != OP_SUCCESS {
		return 0, 0, errorString(r)
	}

	return uint(utilization), uint(samplingPeriodUs), nil
}

func (h handle) DeviceGetEnforcedPowerLimit() (uint, error) {
	var limit C.uint

	r := C.nvmlDeviceGetEnforcedPowerLimit_dlib(h.dev, &limit)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(limit), nil
}

func (h handle) DeviceGetFanSpeed() (uint, error) {
	var speed C.uint

	r := C.nvmlDeviceGetFanSpeed_dlib(h.dev, &speed)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(speed), nil
}

func (h handle) DeviceGetGpuOperationMode() (curMode GpuOperationMode, pendingMode GpuOperationMode, err error) {
	var current, pending C.nvmlGpuOperationMode_t

	r := C.nvmlDeviceGetGpuOperationMode_dlib(h.dev, &current, &pending)

	if r != OP_SUCCESS {
		return GOM_UNKNOWN, GOM_UNKNOWN, errorString(r)
	}

	return gpuOperationMode(current), gpuOperationMode(pending), nil
}

func (h handle) GetGraphicsRunningProcesses(size int) ([]*ProcessInfo, error) {
	var procs = make([]C.nvmlProcessInfo_t, size)
	var count = C.uint(size)

	r := C.nvmlDeviceGetGraphicsRunningProcesses_dlib(h.dev, &count, &procs[0])
	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	n := int(count)
	info := make([]*ProcessInfo, n)
	for i := 0; i < n; i++ {
		info[i] = &ProcessInfo{
			Pid:           uint(procs[i].pid),
			UsedGPUMemory: uint64(procs[i].usedGpuMemory),
		}
	}

	return info, nil
}

func DeviceGetHandleByIndex(idx uint) (handle, error) {
	var dev C.nvmlDevice_t

	r := C.nvmlDeviceGetHandleByIndex_dlib(C.uint(idx), &dev)

	return handle{dev}, errorString(r)
}

func DeviceGetHandleByPciBusId(pciBusID string) (handle, error) {
	var dev C.nvmlDevice_t

	r := C.nvmlDeviceGetHandleByPciBusId_dlib(C.CString(pciBusID), &dev)

	return handle{dev}, errorString(r)
}

func DeviceGetHandleBySerial(serial string) (handle, error) {
	var dev C.nvmlDevice_t

	r := C.nvmlDeviceGetHandleBySerial_dlib(C.CString(serial), &dev)

	return handle{dev}, errorString(r)
}

func DeviceGetHandleByUUID(uuid string) (handle, error) {
	var dev C.nvmlDevice_t

	r := C.nvmlDeviceGetHandleByUUID_dlib(C.CString(uuid), &dev)

	return handle{dev}, errorString(r)
}

func (h handle) DeviceGetIndex() (uint, error) {
	var index C.uint

	r := C.nvmlDeviceGetIndex_dlib(h.dev, &index)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(index), nil
}

func (h handle) DeviceGetInforomConfigurationChecksum() (uint, error) {
	var checksum C.uint

	r := C.nvmlDeviceGetInforomConfigurationChecksum_dlib(h.dev, &checksum)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(checksum), nil
}

func (h handle) DeviceGetInforomImageVersion() (string, error) {
	var version [szName]C.char

	r := C.nvmlDeviceGetInforomImageVersion_dlib(h.dev, &version[0], szName)

	if r != OP_SUCCESS {
		return "", errorString(r)
	}

	return C.GoString(&version[0]), nil
}

func (h handle) DeviceGetInforomVersion(object InforomObject) (string, error) {
	var version [szName]C.char

	r := C.nvmlDeviceGetInforomVersion_dlib(h.dev, object.convert(), &version[0], szName)

	if r != OP_SUCCESS {
		return "", errorString(r)
	}

	return C.GoString(&version[0]), nil
}

func (h handle) DeviceGetMaxClockInfo(clockType ClockType) (uint, error) {
	var clockMHz C.uint

	r := C.nvmlDeviceGetMaxClockInfo_dlib(h.dev, clockType.convert(), &clockMHz)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(clockMHz), nil
}

func (h handle) DeviceGetMaxPcieLinkGeneration() (uint, error) {
	var maxLinkGen C.uint

	r := C.nvmlDeviceGetMaxPcieLinkGeneration_dlib(h.dev, &maxLinkGen)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(maxLinkGen), nil
}

func (h handle) DeviceGetMaxPcieLinkWidth() (uint, error) {
	var maxWidth C.uint

	r := C.nvmlDeviceGetMaxPcieLinkWidth_dlib(h.dev, &maxWidth)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(maxWidth), nil
}

func (h handle) DeviceGetMemoryErrorCounter(mt MemoryErrorType, ec EccCounterType, loc MemoryLocation) (uint64, error) {
	var count C.ulonglong

	r := C.nvmlDeviceGetMemoryErrorCounter_dlib(h.dev, mt.convert(), ec.convert(), loc.convert(), &count)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint64(count), nil
}

func (h handle) DeviceGetMemoryInfo() (free uint64, used uint64, total uint64, err error) {
	var info C.nvmlMemory_t

	r := C.nvmlDeviceGetMemoryInfo_dlib(h.dev, &info)

	if r != OP_SUCCESS {
		return 0, 0, 0, errorString(r)
	}

	return uint64(info.free), uint64(info.used), uint64(info.total), nil
}

func (h handle) DeviceGetMinorNumber() (uint, error) {
	var minor C.uint

	r := C.nvmlDeviceGetMinorNumber_dlib(h.dev, &minor)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(minor), nil
}

func (h handle) DeviceGetMultiGpuBoard() (uint, error) {
	var multi C.uint

	r := C.nvmlDeviceGetMultiGpuBoard_dlib(h.dev, &multi)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(multi), nil
}

func (h handle) DeviceGetName() (string, error) {
	var name [szName]C.char

	r := C.nvmlDeviceGetName_dlib(h.dev, &name[0], szName)

	return C.GoString(&name[0]), errorString(r)
}

func (h handle) DeviceGetPciInfo() (*PciInfo, error) {
	var info C.nvmlPciInfo_t

	r := C.nvmlDeviceGetPciInfo_dlib(h.dev, &info)

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	return &PciInfo{
		BusID:          C.GoString(&info.busId[0]),
		Domain:         uint(info.domain),
		Bus:            uint(info.bus),
		Device:         uint(info.device),
		PciDeviceID:    uint(info.pciDeviceId),
		PciSubSystemID: uint(info.pciSubSystemId),
	}, nil
}

func (h handle) DeviceGetPcieReplayCounter() (uint, error) {
	var value C.uint

	r := C.nvmlDeviceGetPcieReplayCounter_dlib(h.dev, &value)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(value), nil
}

func (h handle) DeviceGetPcieThroughput(counterType PcieUtilCounter) (uint, error) {
	var value C.uint

	r := C.nvmlDeviceGetPcieThroughput_dlib(h.dev, counterType.convert(), &value)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(value), nil
}

func (h handle) DeviceGetPerformanceState() (uint, error) {
	var st C.nvmlPstates_t

	r := C.nvmlDeviceGetPerformanceState_dlib(h.dev, &st)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(C.pstates_to_int(st)), nil
}

func (h handle) DeviceGetPersistenceMode() (bool, error) {
	var state C.nvmlEnableState_t

	r := C.nvmlDeviceGetPersistenceMode_dlib(h.dev, &state)

	if r != OP_SUCCESS {
		return false, errorString(r)
	}

	return stateBool(state), nil
}

func (h handle) DeviceGetPowerManagementDefaultLimit() (uint, error) {
	var defLimit C.uint

	r := C.nvmlDeviceGetPowerManagementDefaultLimit_dlib(h.dev, &defLimit)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(defLimit), nil
}

func (h handle) DeviceGetPowerManagementLimit() (uint, error) {
	var limit C.uint

	r := C.nvmlDeviceGetPowerManagementLimit_dlib(h.dev, &limit)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(limit), nil
}

func (h handle) DeviceGetPowerManagementLimitConstraints() (min uint, max uint, err error) {
	var minLimit, maxLimit C.uint

	r := C.nvmlDeviceGetPowerManagementLimitConstraints_dlib(h.dev, &minLimit, &maxLimit)

	if r != OP_SUCCESS {
		return 0, 0, errorString(r)
	}

	return uint(minLimit), uint(maxLimit), nil
}

func (h handle) DeviceGetPowerManagementMode() (bool, error) {
	var state C.nvmlEnableState_t

	r := C.nvmlDeviceGetPowerManagementMode_dlib(h.dev, &state)

	if r != OP_SUCCESS {
		return false, errorString(r)
	}

	return stateBool(state), nil
}

func (h handle) DeviceGetPowerState() (uint, error) {
	var pstate C.nvmlPstates_t

	r := C.nvmlDeviceGetPowerState_dlib(h.dev, &pstate)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(C.pstates_to_int(pstate)), nil
}

func (h handle) DeviceGetPowerUsage() (uint, error) {
	var power C.uint

	r := C.nvmlDeviceGetPowerUsage_dlib(h.dev, &power)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint(power), nil
}

func (h handle) DeviceGetRetiredPages(cause PageRetirementCause) ([]uint64, error) {
	var (
		pageCount   C.uint
		peekAddress C.ulonglong
		addrs       []uint64
	)

	pageCount = C.uint(0)
	r := C.nvmlDeviceGetRetiredPages_dlib(h.dev, cause.convert(), &pageCount, &peekAddress)
	if r != OP_SUCCESS && r != OP_INSUFFICIENT_SIZE {
		return nil, errorString(r)
	}

	if pageCount == 0 {
		return nil, nil
	}

	addresses := make([]C.ulonglong, uint(pageCount))
	r = C.nvmlDeviceGetRetiredPages_dlib(h.dev, cause.convert(), &pageCount, &addresses[0])

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	for i := 0; i < int(pageCount); i++ {
		addrs = append(addrs, uint64(addresses[i]))
	}

	return addrs, nil
}

func (h handle) DeviceGetRetiredPagesPendingStatus() (bool, error) {
	var state C.nvmlEnableState_t

	r := C.nvmlDeviceGetRetiredPagesPendingStatus_dlib(h.dev, &state)

	if r != OP_SUCCESS {
		return false, errorString(r)
	}

	return stateBool(state), nil
}

func (h handle) DeviceGetSerial() (string, error) {
	var serial [szUUID]C.char

	r := C.nvmlDeviceGetSerial_dlib(h.dev, &serial[0], C.uint(szUUID))

	return C.GoString(&serial[0]), errorString(r)
}

func (h handle) DeviceGetSupportedClocksThrottleReasons() (uint64, error) {
	var reason C.ulonglong

	r := C.nvmlDeviceGetSupportedClocksThrottleReasons_dlib(h.dev, &reason)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint64(reason), nil
}

func (h handle) DeviceGetSupportedGraphicsClocks(memoryClockMHz uint) ([]uint, error) {
	var (
		count     C.uint
		peekArray C.uint
		clocks    []uint
	)

	count = C.uint(0)
	r := C.nvmlDeviceGetSupportedGraphicsClocks_dlib(h.dev, C.uint(memoryClockMHz), &count, &peekArray)

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	d := make([]C.uint, uint(count))
	r = C.nvmlDeviceGetSupportedGraphicsClocks_dlib(h.dev, C.uint(memoryClockMHz), &count, &d[0])

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	for i := 0; i < int(count); i++ {
		clocks = append(clocks, uint(d[i]))
	}

	return clocks, nil
}

func (h handle) DeviceGetSupportedMemoryClocks() ([]uint, error) {
	var (
		count     C.uint
		peekArray C.uint
		clocks    []uint
	)

	count = C.uint(0)
	r := C.nvmlDeviceGetSupportedMemoryClocks_dlib(h.dev, &count, &peekArray)

	if r != OP_SUCCESS && r != OP_INSUFFICIENT_SIZE {
		return nil, errorString(r)
	}

	if count == 0 {
		return nil, nil
	}

	d := make([]C.uint, uint(count))

	r = C.nvmlDeviceGetSupportedMemoryClocks_dlib(h.dev, &count, &d[0])

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	for i := 0; i < int(count); i++ {
		clocks = append(clocks, uint(d[i]))
	}

	return clocks, nil
}

func (h handle) DeviceGetTemperature() (uint, error) {
	var temp C.uint

	r := C.nvmlDeviceGetTemperature_dlib(h.dev, C.NVML_TEMPERATURE_GPU, &temp)

	return uint(temp), errorString(r)
}

func (h handle) DeviceGetTemperatureThreshold(threshold TemperatureThresholds) (uint, error) {
	var temp C.uint

	r := C.nvmlDeviceGetTemperatureThreshold_dlib(h.dev, threshold.convert(), &temp)

	return uint(temp), errorString(r)
}

func DeviceGetTopologyCommonAncestor(h1, h2 handle) (GpuTopologyLevel, error) {
	var ancestor C.nvmlGpuTopologyLevel_t

	r := C.nvmlDeviceGetTopologyCommonAncestor_dlib(h1.dev, h2.dev, &ancestor)

	return GpuTopologyLevel(C.gpuTopologyLevel_to_int(ancestor)), errorString(r)
}

func (h handle) DeviceGetTopologyNearestGpus(level GpuTopologyLevel) ([]handle, error) {
	var (
		count C.uint
		devs  []handle
	)

	d := make([]C.nvmlDevice_t, maxDevices)
	count = maxDevices

	r := C.nvmlDeviceGetTopologyNearestGpus_dlib(h.dev, level.convert(), &count, &d[0])

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	for i := 0; i < int(count); i++ {
		devs = append(devs, handle{d[i]})
	}

	return devs, nil
}

func (h handle) DeviceGetTotalEccErrors(mt MemoryErrorType, et EccCounterType) (uint64, error) {
	var eccCount C.ulonglong

	r := C.nvmlDeviceGetTotalEccErrors_dlib(h.dev, mt.convert(), et.convert(), &eccCount)

	if r != OP_SUCCESS {
		return 0, errorString(r)
	}

	return uint64(eccCount), nil
}

func (h handle) DeviceGetUUID() (string, error) {
	var uuid [szUUID]C.char

	r := C.nvmlDeviceGetUUID_dlib(h.dev, &uuid[0], C.uint(szUUID))

	if r != OP_SUCCESS {
		return "", errorString(r)
	}

	return C.GoString(&uuid[0]), nil
}

func (h handle) DeviceGetUtilizationRates() (*Utilization, error) {
	var util C.nvmlUtilization_t

	r := C.nvmlDeviceGetUtilizationRates_dlib(h.dev, &util)

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	return &Utilization{
		GPU:    uint(util.gpu),
		Memory: uint(util.memory),
	}, nil
}

func (h handle) DeviceGetVbiosVersion() (string, error) {
	var ver [szName]C.char

	r := C.nvmlDeviceGetVbiosVersion_dlib(h.dev, &ver[0], C.uint(szName))

	if r != OP_SUCCESS {
		return "", errorString(r)
	}

	return C.GoString(&ver[0]), nil
}

func (h handle) DeviceGetViolationStatus(policy PerfPolicy) (*ViolationTime, error) {
	var d C.nvmlViolationTime_t

	r := C.nvmlDeviceGetViolationStatus_dlib(h.dev, policy.convert(), &d)

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	return &ViolationTime{
		ReferTime:     time.Duration(d.referenceTime),
		ViolationTime: time.Duration(d.violationTime),
	}, nil
}

func DeviceOnSameBoard(h1, h2 handle) (bool, error) {
	var d C.int

	r := C.nvmlDeviceOnSameBoard_dlib(h1.dev, h2.dev, &d)

	if r != OP_SUCCESS {
		return false, errorString(r)
	}

	test := int(d)

	if test != 0 {
		return true, nil
	}

	return false, nil
}

func (h handle) DeviceResetApplicationsClocks() error {
	r := C.nvmlDeviceResetApplicationsClocks_dlib(h.dev)

	return errorString(r)
}

func (h handle) DeviceSetAutoBoostedClocksEnabled() (bool, error) {
	var state C.nvmlEnableState_t

	r := C.nvmlDeviceSetAutoBoostedClocksEnabled_dlib(h.dev, state)

	if r != OP_SUCCESS {
		return false, errorString(r)
	}

	return stateBool(state), nil
}

func (h handle) DeviceSetCpuAffinity() error {
	r := C.nvmlDeviceSetCpuAffinity_dlib(h.dev)

	return errorString(r)
}

func (h handle) DeviceSetDefaultAutoBoostedClocksEnabled(enabled bool) error {
	var flags C.uint

	r := C.nvmlDeviceSetDefaultAutoBoostedClocksEnabled_dlib(h.dev, boolState(enabled), flags)

	return errorString(r)
}

func (h handle) DeviceValidateInforom() error {
	r := C.nvmlDeviceValidateInforom_dlib(h.dev)

	return errorString(r)
}

func SystemGetTopologyGpuSet(cpu int) ([]handle, error) {
	var (
		count C.uint
		d     []C.nvmlDevice_t
		devs  []handle
	)

	count = C.uint(1)
	for {
		d = make([]C.nvmlDevice_t, uint(count))

		r := C.nvmlSystemGetTopologyGpuSet_dlib(C.uint(cpu), &count, &d[0])

		if r == C.NVML_ERROR_INVALID_ARGUMENT {
			count++
			continue
		}

		if r != OP_SUCCESS {
			return nil, errorString(r)
		}

		break
	}

	for i := 0; i < int(count); i++ {
		devs = append(devs, handle{d[i]})
	}

	return devs, nil
}

func SystemGetHicVersion() ([]*HwbcEntry, error) {
	var (
		count   C.uint
		entries []*HwbcEntry
	)

	count = C.uint(maxDevices)
	d := make([]C.nvmlHwbcEntry_t, uint(count))

	r := C.nvmlSystemGetHicVersion_dlib(&count, &d[0])

	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	for i := 0; i < int(count); i++ {
		entries = append(entries, &HwbcEntry{
			ID:        uint(d[i].hwbcId),
			FwVersion: C.GoString(&d[i].firmwareVersion[0]),
		})
	}

	return entries, nil
}

func (h handle) DeviceClearEccErrorCounts(counterType EccCounterType) error {
	r := C.nvmlDeviceClearEccErrorCounts_dlib(h.dev, counterType.convert())

	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func (h handle) DeviceSetAPIRestriction(api RestrictedAPI, isEnabled bool) error {
	r := C.nvmlDeviceSetAPIRestriction_dlib(h.dev, api.convert(), boolState(isEnabled))

	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func (h handle) DeviceSetApplicationsClocks(memClocksMHz, clocksMHz uint) error {
	r := C.nvmlDeviceSetApplicationsClocks_dlib(h.dev, C.uint(memClocksMHz), C.uint(clocksMHz))

	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func (h handle) DeviceSetComputeMode(mode ComputeMode) error {
	r := C.nvmlDeviceSetComputeMode_dlib(h.dev, mode.convert())

	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func (h handle) DeviceSetEccMode(isEnabled bool) error {
	r := C.nvmlDeviceSetEccMode_dlib(h.dev, boolState(isEnabled))

	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func (h handle) DeviceSetGpuOperationMode(mode GpuOperationMode) error {
	r := C.nvmlDeviceSetGpuOperationMode_dlib(h.dev, mode.convert())

	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func (h handle) DeviceSetPersistenceMode(isEnabled bool) error {
	r := C.nvmlDeviceSetPersistenceMode_dlib(h.dev, boolState(isEnabled))

	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func (h handle) DeviceSetPowerManagementLimit(limit uint) error {
	r := C.nvmlDeviceSetPowerManagementLimit_dlib(h.dev, C.uint(limit))

	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func (h handle) DeviceGetAverageGPUUsage(since time.Duration) (uint, error) {
	lastTs := C.ulonglong(time.Now().Add(-1*since).UnixNano() / 1000)
	var n C.uint
	r := C.__nvmlDeviceGetAverageUsage(h.dev, C.NVML_GPU_UTILIZATION_SAMPLES, lastTs, &n)
	return uint(n), errorString(r)
}

func (h handle) DeviceGetProcessUtilization(maxProcess int, since time.Duration) ([]*ProcessUtilizationSample, error) {
	lastTs := C.ulonglong(time.Now().Add(-1*since).UnixNano() / 1000)
	var (
		count   C.uint
		samples []*ProcessUtilizationSample
	)

	count = C.uint(maxProcess)
	d := make([]C.nvmlProcessUtilizationSample_t, uint(count))

	r := C.nvmlDeviceGetProcessUtilization_dlib(h.dev, &d[0], &count, lastTs)

	if r != OP_SUCCESS && r != C.NVML_ERROR_NOT_FOUND {
		return nil, errorString(r)
	}

	for i := 0; i < int(count); i++ {
		samples = append(samples, &ProcessUtilizationSample{
			Pid:       uint(d[i].pid),
			TimeStamp: time.Duration(d[i].timeStamp),
			SmUtil:    uint(d[i].smUtil),
			MemUtil:   uint(d[i].memUtil),
			EncUtil:   uint(d[i].encUtil),
			DecUtil:   uint(d[i].decUtil),
		})
	}

	return samples, nil
}

func (h handle) DeviceGetSupportedEventTypes() ([]EventType, error) {
	var supportedType C.ulonglong
	r := C.nvmlDeviceGetSupportedEventTypes_dlib(h.dev, &supportedType)
	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	return eventTypes(supportedType), nil
}

func (h handle) DeviceRegisterEvents(evtType EventType, set EventSet) error {
	r := C.nvmlDeviceRegisterEvents_dlib(h.dev, C.ulonglong(evtType), set.set)
	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func EventSetCreate() (*EventSet, error) {
	var set C.nvmlEventSet_t

	r := C.nvmlEventSetCreate_dlib(&set)
	if r != OP_SUCCESS {
		return nil, errorString(r)
	}

	return &EventSet{set: set}, nil
}

func EventSetFree(set *EventSet) error {
	r := C.nvmlEventSetFree_dlib(set.set)

	if r != OP_SUCCESS {
		return errorString(r)
	}

	return nil
}

func EventSetWait(set EventSet, timeoutMS int) (*EventData, error) {
	var data C.nvmlEventData_t

	r := C.nvmlEventSetWait_dlib(set.set, &data, C.uint(timeoutMS))

	if r != OP_SUCCESS && r != OP_TIMEOUT {
		return nil, errorString(r)
	}

	if r == OP_TIMEOUT {
		return nil, nil
	}

	return &EventData{
		Device: handle{data.device},
		Data:   uint64(data.eventData),
		Types:  eventTypes(data.eventType),
	}, nil
}
