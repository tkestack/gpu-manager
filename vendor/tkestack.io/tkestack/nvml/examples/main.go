/*
 * Copyright 2019 THL A29 Limited, a Tencent company.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"time"

	"tkestack.io/tkestack/nvml"
)

func failedMsg(msg string, err error) {
	fmt.Printf("%s: %+v\n", msg, err)
}

func main() {
	if err := nvml.Init(); err != nil {
		fmt.Printf("nvml error: %+v", err)
		return
	}
	defer nvml.Shutdown()

	ver, err := nvml.SystemGetDriverVersion()
	if err != nil {
		failedMsg("SystemGetDriverVersion", err)
	} else {
		fmt.Printf("SystemGetDriverVersion: %s\n", ver)
	}

	nvVer, err := nvml.SystemGetNVMLVersion()
	if err != nil {
		failedMsg("SystemGetNVMLVersion", err)
	} else {
		fmt.Printf("nv driver: %s\n", nvVer)
	}

	name, err := nvml.SystemGetProcessName(1)
	if err != nil {
		failedMsg("SystemGetProcessName", err)
	} else {
		fmt.Printf("No.1 process name is %s\n", name)
	}

	num, err := nvml.DeviceGetCount()
	if err != nil {
		failedMsg("DeviceGetCount", err)
	} else {
		fmt.Printf("We have %d cards\n", num)
	}

	cmpDev, _ := nvml.DeviceGetHandleByIndex(0)

	for i := uint(0); i < num; i++ {
		fmt.Println("============")

		dev, err := nvml.DeviceGetHandleByIndex(i)
		if err != nil {
			failedMsg("DeviceGetHandleByIndex", err)
		} else {
			fmt.Printf("Get dev %d\n", i)
		}

		err = dev.DeviceClearCpuAffinity()
		if err != nil {
			failedMsg("DeviceClearCpuAffinity", err)
		} else {
			fmt.Printf("Clear DeviceClearCpuAffinity\n")
		}

		enabled, err := dev.DeviceGetAPIRestriction(nvml.RESTRICTED_API_SET_APPLICATION_CLOCKS)
		if err != nil {
			failedMsg("DeviceGetAPIRestriction", err)
		} else {
			fmt.Printf("RESTRICTED_API_SET_APPLICATION_CLOCKS: %+v\n", enabled)
		}

		clockMHz, err := dev.DeviceGetApplicationsClock(nvml.CLOCK_GRAPHICS)
		if err != nil {
			failedMsg("DeviceGetApplicationsClock", err)
		} else {
			fmt.Printf("DeviceGetApplicationsClock: %d\n", clockMHz)
		}

		boost, defVal, err := dev.DeviceGetAutoBoostedClocksEnabled()
		if err != nil {
			failedMsg("DeviceGetAutoBoostedClocksEnabled", err)
		} else {
			fmt.Printf("DeviceGetAutoBoostedClocksEnabled: %+v, %+v\n", boost, defVal)
		}

		free, total, used, err := dev.DeviceGetBAR1MemoryInfo()
		if err != nil {
			failedMsg("DeviceGetBAR1MemoryInfo", err)
		} else {
			fmt.Printf("DeviceGetBAR1MemoryInfo: free: %d, total: %d, used: %d\n", free, used, total)
		}

		boardID, err := dev.DeviceGetBoardId()
		if err != nil {
			failedMsg("DeviceGetBoardId", err)
		} else {
			fmt.Printf("DeviceGetBoardId: %d\n", boardID)
		}

		brandType, err := dev.DeviceGetBrand()
		if err != nil {
			failedMsg("DeviceGetBrand", err)
		} else {
			fmt.Printf("DeviceGetBrand: %+v\n", brandType)
		}

		bridgeChipInfos, err := dev.DeviceGetBridgeChipInfo()
		if err != nil {
			failedMsg("DeviceGetBridgeChipInfo", err)
		} else {
			fmt.Printf("DeviceGetBridgeChipInfo: %d\n", len(bridgeChipInfos))

			for _, info := range bridgeChipInfos {
				fmt.Printf("\tver: %d, type: %d\n", info.FwVersion, info.Type)
			}
		}

		mode, err := dev.DeviceGetComputeMode()
		if err != nil {
			failedMsg("DeviceGetComputeMode", err)
		} else {
			fmt.Printf("DeviceGetComputeMode: %d\n", mode)
		}

		processes, err := dev.DeviceGetComputeRunningProcesses(32)
		if err != nil {
			failedMsg("DeviceGetComputeRunningProcesses", err)
		} else {
			fmt.Printf("DeviceGetComputeRunningProcesses: %d\n", len(processes))
			for _, proc := range processes {
				fmt.Printf("\tpid: %d, usedMemory: %d", proc.Pid, proc.UsedGPUMemory)
			}
		}

		cpusets, err := dev.DeviceGetCpuAffinity(2)
		if err != nil {
			failedMsg("DeviceGetCpuAffinity", err)
		} else {
			fmt.Printf("DeviceGetCpuAffinity: %+v\n", cpusets)
		}

		linkgen, err := dev.DeviceGetCurrPcieLinkGeneration()
		if err != nil {
			failedMsg("DeviceGetCurrPcieLinkGeneration", err)
		} else {
			fmt.Printf("DeviceGetCurrPcieLinkGeneration: %d\n", linkgen)
		}

		width, err := dev.DeviceGetCurrPcieLinkWidth()
		if err != nil {
			failedMsg("DeviceGetCurrPcieLinkWidth", err)
		} else {
			fmt.Printf("DeviceGetCurrPcieLinkWidth: %d\n", width)
		}

		reasons, err := dev.DeviceGetCurrentClocksThrottleReasons()
		if err != nil {
			failedMsg("DeviceGetCurrentClocksThrottleReasons", err)
		} else {
			fmt.Printf("DeviceGetCurrentClocksThrottleReasons: %d\n", len(reasons))
			for _, reason := range reasons {
				fmt.Printf("\tReason: %+v\n", reason)
			}
		}

		decodeUtil, decodePeriod, err := dev.DeviceGetDecoderUtilization()
		if err != nil {
			failedMsg("DeviceGetDecoderUtilization", err)
		} else {
			fmt.Printf("DeviceGetDecoderUtilization: %d, %d\n", decodeUtil, decodePeriod)
		}

		defClock, err := dev.DeviceGetDefaultApplicationsClock(nvml.CLOCK_GRAPHICS)
		if err != nil {
			failedMsg("DeviceGetDefaultApplicationsClock", err)
		} else {
			fmt.Printf("DeviceGetDefaultApplicationsClock: %d\n", defClock)
		}

		eccCounter, err := dev.DeviceGetDetailedEccErrors(nvml.MEMORY_ERROR_TYPE_CORRECTED, nvml.VOLATILE_ECC)
		if err != nil {
			failedMsg("DeviceGetDetailedEccErrors", err)
		} else {
			fmt.Printf("DeviceGetDetailedEccErrors: %+v\n", eccCounter)
		}

		display, err := dev.DeviceGetDisplayMode()
		if err != nil {
			failedMsg("DeviceGetDisplayMode", err)
		} else {
			fmt.Printf("DeviceGetDisplayMode: %+v\n", display)
		}

		eccCurrent, eccPending, err := dev.DeviceGetEccMode()
		if err != nil {
			failedMsg("DeviceGetEccMode", err)
		} else {
			fmt.Printf("DeviceGetEccMode: %+v, %+v\n", eccCurrent, eccPending)
		}

		encodeUtil, encodePeriod, err := dev.DeviceGetEncoderUtilization()
		if err != nil {
			failedMsg("DeviceGetEncoderUtilization", err)
		} else {
			fmt.Printf("DeviceGetEncoderUtilization: %d, %d\n", encodeUtil, encodePeriod)
		}

		powerLimit, err := dev.DeviceGetEnforcedPowerLimit()
		if err != nil {
			failedMsg("DeviceGetEnforcedPowerLimit", err)
		} else {
			fmt.Printf("DeviceGetEnforcedPowerLimit: %d\n", powerLimit)
		}

		speed, err := dev.DeviceGetFanSpeed()
		if err != nil {
			failedMsg("DeviceGetFanSpeed", err)
		} else {
			fmt.Printf("DeviceGetFanSpeed: %d\n", speed)
		}

		gpuOpModeCurrent, gpuOpModePending, err := dev.DeviceGetGpuOperationMode()
		if err != nil {
			failedMsg("DeviceGetGpuOperationMode", err)
		} else {
			fmt.Printf("DeviceGetGpuOperationMode: %+v, %+v\n", gpuOpModeCurrent, gpuOpModePending)
		}

		gRunningProcs, err := dev.GetGraphicsRunningProcesses(10)
		if err != nil {
			failedMsg("GetGraphicsRunningProcesses", err)
		} else {
			fmt.Printf("GetGraphicsRunningProcesses: %d\n", len(gRunningProcs))

			for _, proc := range gRunningProcs {
				fmt.Printf("\t%d %d\n", proc.Pid, proc.UsedGPUMemory)
			}
		}

		checkSum, err := dev.DeviceGetInforomConfigurationChecksum()
		if err != nil {
			failedMsg("DeviceGetInforomConfigurationChecksum", err)
		} else {
			fmt.Printf("DeviceGetInforomConfigurationChecksum: %d\n", checkSum)
		}

		inforomImageVer, err := dev.DeviceGetInforomImageVersion()
		if err != nil {
			failedMsg("DeviceGetInforomImageVersion", err)
		} else {
			fmt.Printf("DeviceGetInforomImageVersion: %s\n", inforomImageVer)
		}

		inforomVer, err := dev.DeviceGetInforomVersion(nvml.INFOROM_OEM)
		if err != nil {
			failedMsg("DeviceGetInforomVersion", err)
		} else {
			fmt.Printf("DeviceGetInforomVersion: %s\n", inforomVer)
		}

		maxClock, err := dev.DeviceGetMaxClockInfo(nvml.CLOCK_MEM)
		if err != nil {
			failedMsg("DeviceGetMaxClockInfo", err)
		} else {
			fmt.Printf("DeviceGetMaxClockInfo: %d\n", maxClock)
		}

		maxLinkGen, err := dev.DeviceGetMaxPcieLinkGeneration()
		if err != nil {
			failedMsg("DeviceGetMaxPcieLinkGeneration", err)
		} else {
			fmt.Printf("DeviceGetMaxPcieLinkGeneration: %d\n", maxLinkGen)
		}

		maxWidth, err := dev.DeviceGetMaxPcieLinkWidth()
		if err != nil {
			failedMsg("DeviceGetMaxPcieLinkWidth", err)
		} else {
			fmt.Printf("DeviceGetMaxPcieLinkWidth: %d\n", maxWidth)
		}

		memEccCounter, err := dev.DeviceGetMemoryErrorCounter(
			nvml.MEMORY_ERROR_TYPE_CORRECTED, nvml.VOLATILE_ECC, nvml.MEMORY_LOCATION_DEVICE_MEMORY)
		if err != nil {
			failedMsg("DeviceGetMemoryErrorCounter", err)
		} else {
			fmt.Printf("DeviceGetMemoryErrorCounter: %d\n", memEccCounter)
		}

		memFree, memUsed, memTotal, err := dev.DeviceGetMemoryInfo()
		if err != nil {
			failedMsg("DeviceGetMemoryInfo", err)
		} else {
			fmt.Printf("DeviceGetMemoryInfo: %d, %d, %d\n", memFree, memUsed, memTotal)
		}

		minor, err := dev.DeviceGetMinorNumber()
		if err != nil {
			failedMsg("DeviceGetMinorNumber", err)
		} else {
			fmt.Printf("DeviceGetMinorNumber: %d\n", minor)
		}

		multi, err := dev.DeviceGetMultiGpuBoard()
		if err != nil {
			failedMsg("DeviceGetMultiGpuBoard", err)
		} else {
			fmt.Printf("DeviceGetMultiGpuBoard: %d\n", multi)
		}

		name, err := dev.DeviceGetName()
		if err != nil {
			failedMsg("DeviceGetName", err)
		} else {
			fmt.Printf("DeviceGetName: %s\n", name)
		}

		pciInfo, err := dev.DeviceGetPciInfo()
		if err != nil {
			failedMsg("DeviceGetPciInfo", err)
		} else {
			fmt.Printf("DeviceGetPciInfo: %+v\n", pciInfo)
		}

		replayCounter, err := dev.DeviceGetPcieReplayCounter()
		if err != nil {
			failedMsg("DeviceGetPcieReplayCounter", err)
		} else {
			fmt.Printf("DeviceGetPcieReplayCounter: %d\n", replayCounter)
		}

		throughput, err := dev.DeviceGetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES)
		if err != nil {
			failedMsg("DeviceGetPcieThroughput", err)
		} else {
			fmt.Printf("DeviceGetPcieThroughput: %d\n", throughput)
		}

		performState, err := dev.DeviceGetPerformanceState()
		if err != nil {
			failedMsg("DeviceGetPerformanceState", err)
		} else {
			fmt.Printf("DeviceGetPerformanceState: %d\n", performState)
		}

		persistenceMode, err := dev.DeviceGetPersistenceMode()
		if err != nil {
			failedMsg("DeviceGetPersistenceMode", err)
		} else {
			fmt.Printf("DeviceGetPersistenceMode: %+v\n", persistenceMode)
		}

		powerManagementDefLimit, err := dev.DeviceGetPowerManagementDefaultLimit()
		if err != nil {
			failedMsg("DeviceGetPowerManagementDefaultLimit", err)
		} else {
			fmt.Printf("DeviceGetPowerManagementDefaultLimit: %d\n", powerManagementDefLimit)
		}

		powerManagementLimit, err := dev.DeviceGetPowerManagementLimit()
		if err != nil {
			failedMsg("DeviceGetPowerManagementLimit", err)
		} else {
			fmt.Printf("DeviceGetPowerManagementLimit: %d\n", powerManagementLimit)
		}

		minLimit, maxLimit, err := dev.DeviceGetPowerManagementLimitConstraints()
		if err != nil {
			failedMsg("DeviceGetPowerManagementLimitConstraints", err)
		} else {
			fmt.Printf("DeviceGetPowerManagementLimitConstraints: %d, %d\n", minLimit, maxLimit)
		}

		powerManagementMode, err := dev.DeviceGetPowerManagementMode()
		if err != nil {
			failedMsg("DeviceGetPowerManagementMode", err)
		} else {
			fmt.Printf("DeviceGetPowerManagementMode: %+v\n", powerManagementMode)
		}

		powerState, err := dev.DeviceGetPowerState()
		if err != nil {
			failedMsg("DeviceGetPowerState", err)
		} else {
			fmt.Printf("DeviceGetPowerState: %d\n", powerState)
		}

		powerUsage, err := dev.DeviceGetPowerUsage()
		if err != nil {
			failedMsg("DeviceGetPowerUsage", err)
		} else {
			fmt.Printf("DeviceGetPowerUsage: %d\n", powerUsage)
		}

		addresses, err := dev.DeviceGetRetiredPages(nvml.PAGE_RETIREMENT_CAUSE_DOUBLE_BIT_ECC_ERROR)
		if err != nil {
			failedMsg("DeviceGetRetiredPages", err)
		} else {
			fmt.Printf("DeviceGetRetiredPages: %d\n", len(addresses))
			for _, addr := range addresses {
				fmt.Printf("\t %d\n", addr)
			}
		}

		pendingStatus, err := dev.DeviceGetRetiredPagesPendingStatus()
		if err != nil {
			failedMsg("DeviceGetRetiredPagesPendingStatus", err)
		} else {
			fmt.Printf("DeviceGetRetiredPagesPendingStatus: %+v\n", pendingStatus)
		}

		serial, err := dev.DeviceGetSerial()
		if err != nil {
			failedMsg("DeviceGetSerial", err)
		} else {
			fmt.Printf("DeviceGetSerial: %s\n", serial)
		}

		CTReason, err := dev.DeviceGetSupportedClocksThrottleReasons()
		if err != nil {
			failedMsg("DeviceGetSupportedClocksThrottleReasons", err)
		} else {
			fmt.Printf("DeviceGetSupportedClocksThrottleReasons: %d\n", CTReason)
		}

		sGCs, err := dev.DeviceGetSupportedGraphicsClocks(1000)
		if err != nil {
			failedMsg("DeviceGetSupportedGraphicsClocks", err)
		} else {
			fmt.Printf("DeviceGetSupportedGraphicsClocks: %d\n", len(sGCs))
			for _, d := range sGCs {
				fmt.Printf("\tReason: %d\n", d)
			}
		}

		sMCs, err := dev.DeviceGetSupportedMemoryClocks()
		if err != nil {
			failedMsg("DeviceGetSupportedMemoryClocks", err)
		} else {
			fmt.Printf("DeviceGetSupportedMemoryClocks: %d\n", len(sMCs))
			for _, d := range sMCs {
				fmt.Printf("\tClocks: %d\n", d)
			}
		}

		temper, err := dev.DeviceGetTemperature()
		if err != nil {
			failedMsg("DeviceGetTemperature", err)
		} else {
			fmt.Printf("DeviceGetTemperature: %d\n", temper)
		}

		temperThreshold, err := dev.DeviceGetTemperatureThreshold(nvml.TEMPERATURE_THRESHOLD_SLOWDOWN)
		if err != nil {
			failedMsg("DeviceGetTemperatureThreshold", err)
		} else {
			fmt.Printf("DeviceGetTemperatureThreshold: %d\n", temperThreshold)
		}

		gpuLevel, err := nvml.DeviceGetTopologyCommonAncestor(cmpDev, dev)
		if err != nil {
			failedMsg("DeviceGetTopologyCommonAncestor", err)
		} else {
			fmt.Printf("DeviceGetTopologyCommonAncestor: %d\n", gpuLevel)
		}

		nearest, err := dev.DeviceGetTopologyNearestGpus(nvml.TOPOLOGY_HOSTBRIDGE)
		if err != nil {
			failedMsg("DeviceGetTopologyNearestGpus", err)
		} else {
			fmt.Printf("DeviceGetTopologyNearestGpus: %+v\n", nearest)
		}

		tCounter, err := dev.DeviceGetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_UNCORRECTED, nvml.VOLATILE_ECC)
		if err != nil {
			failedMsg("DeviceGetTotalEccErrors", err)
		} else {
			fmt.Printf("DeviceGetTotalEccErrors: volatile %d\n", tCounter)
		}

		tCounter, err = dev.DeviceGetTotalEccErrors(nvml.MEMORY_ERROR_TYPE_UNCORRECTED, nvml.AGGREGATE_ECC)
		if err != nil {
			failedMsg("DeviceGetTotalEccErrors", err)
		} else {
			fmt.Printf("DeviceGetTotalEccErrors: agg %d\n", tCounter)
		}

		uuid, err := dev.DeviceGetUUID()
		if err != nil {
			failedMsg("DeviceGetUUID", err)
		} else {
			fmt.Printf("DeviceGetUUID: %s\n", uuid)
		}

		util, err := dev.DeviceGetUtilizationRates()
		if err != nil {
			failedMsg("DeviceGetUtilizationRates", err)
		} else {
			fmt.Printf("DeviceGetUtilizationRates: %+v\n", util)
		}

		vbios, err := dev.DeviceGetVbiosVersion()
		if err != nil {
			failedMsg("DeviceGetVbiosVersion", err)
		} else {
			fmt.Printf("DeviceGetVbiosVersion: %s\n", vbios)
		}

		violation, err := dev.DeviceGetViolationStatus(nvml.PERF_POLICY_POWER)
		if err != nil {
			failedMsg("DeviceGetViolationStatus", err)
		} else {
			fmt.Printf("DeviceGetViolationStatus: %+v\n", violation)
		}

		sameBoard, err := nvml.DeviceOnSameBoard(cmpDev, dev)
		if err != nil {
			failedMsg("DeviceOnSameBoard", err)
		} else {
			fmt.Printf("DeviceOnSameBoard: %+v\n", sameBoard)
		}

		if err := dev.DeviceResetApplicationsClocks(); err != nil {
			failedMsg("DeviceResetApplicationsClocks", err)
		}

		autoBoost, err := dev.DeviceSetAutoBoostedClocksEnabled()
		if err != nil {
			failedMsg("DeviceSetAutoBoostedClocksEnabled", err)
		} else {
			fmt.Printf("DeviceSetAutoBoostedClocksEnabled: %+v\n", autoBoost)
		}

		if err := dev.DeviceSetCpuAffinity(); err != nil {
			failedMsg("DeviceSetCpuAffinity", err)
		}

		if err := dev.DeviceSetDefaultAutoBoostedClocksEnabled(true); err != nil {
			failedMsg("DeviceSetDefaultAutoBoostedClocksEnabled", err)
		}

		if err := dev.DeviceValidateInforom(); err != nil {
			failedMsg("DeviceValidateInforom", err)
		}

		affinity, err := nvml.SystemGetTopologyGpuSet(0)
		if err != nil {
			failedMsg("SystemGetTopologyGpuSet", err)
		} else {
			fmt.Printf("SystemGetTopologyGpuSet: %+v\n", affinity)
		}

		processSamples, err := dev.DeviceGetProcessUtilization(128, time.Second)
		if err != nil {
			failedMsg("DeviceGetProcessUtilization", err)
		}

		fmt.Printf("DeviceGetProcessUtilization: %v\n", processSamples)

		func() {
			var typeMask nvml.EventType = nvml.EventTypeNone
			supportedTypes, err := dev.DeviceGetSupportedEventTypes()
			if err != nil {
				failedMsg("DeviceGetSupportedEventTypes", err)
				return
			}

			fmt.Printf("Support Type: \n")
			for _, t := range supportedTypes {
				typeMask |= t
				fmt.Printf("\t%d\n", t)
			}

			evtSet, err := nvml.EventSetCreate()
			if err != nil {
				failedMsg("EventSetCreate", err)
				return
			}

			fmt.Printf("EventSetCreate\n")

			defer func() {
				deferErr := nvml.EventSetFree(evtSet)
				if deferErr != nil {
					failedMsg("EventSetFree", deferErr)
				} else {
					fmt.Printf("EventSetFree\n")
				}
			}()

			err = dev.DeviceRegisterEvents(typeMask, *evtSet)
			if err != nil {
				failedMsg("DeviceRegisterEvents", err)
			} else {
				fmt.Printf("DeviceRegisterEvents\n")
			}

			result, wErr := nvml.EventSetWait(*evtSet, 10000)
			if wErr != nil {
				failedMsg("EventSetWait", wErr)
			} else {
				if result == nil {
					fmt.Printf("Timeout: no event\n")
					return
				}

				fmt.Printf("EventSetWait: \n")
				for _, t := range result.Types {
					fmt.Printf("Type: %d\n", t)
				}
				fmt.Printf("Data: %d\n", result.Data)
				uuid, _ := result.Device.DeviceGetUUID()
				fmt.Printf("Device: %s\n", uuid)
			}
		}()
	}

	hwcb, err := nvml.SystemGetHicVersion()
	if err != nil {
		failedMsg("SystemGetHicVersion", err)
	} else {
		fmt.Printf("SystemGetHicVersion: %+v\n", hwcb)
	}
}
