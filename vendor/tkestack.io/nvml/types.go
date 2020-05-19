// +build !windows

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

package nvml

// #cgo CFLAGS: -I/usr/local/cuda/include
// #include "nvml_dl.h"
import "C"
import "time"

const (
	szDriver             = C.NVML_SYSTEM_DRIVER_VERSION_BUFFER_SIZE
	szName               = C.NVML_DEVICE_NAME_BUFFER_SIZE
	szUUID               = C.NVML_DEVICE_UUID_BUFFER_SIZE
	szProcName           = 64
	OP_SUCCESS           = C.NVML_SUCCESS
	OP_INSUFFICIENT_SIZE = C.NVML_ERROR_INSUFFICIENT_SIZE
	OP_TIMEOUT           = C.NVML_ERROR_TIMEOUT
	maxDevices           = 128
)

type RestrictedAPI int

const (
	RESTRICTED_API_SET_APPLICATION_CLOCKS RestrictedAPI = iota
	RESTRICTED_API_SET_AUTO_BOOSTED_CLOCKS
	RESTRICTED_API_COUNT
)

type ClockType int

const (
	CLOCK_GRAPHICS ClockType = iota
	CLOCK_SM
	CLOCK_MEM
	CLOCK_COUNT
)

type BrandType int

const (
	BRAND_UNKNOWN BrandType = iota
	BRAND_QUADRO
	BRAND_TESLA
	BRAND_NVS
	BRAND_GRID
	BRAND_GEFORCE
	BRAND_COUNT
)

type handle struct{ dev C.nvmlDevice_t }

type BridgeChipInfo struct {
	FwVersion uint
	Type      BridgeChipType
}

type BridgeChipType int

const (
	BRIDGE_CHIP_PLX BridgeChipType = iota
	BRIDGE_CHIP_BRO4
)

type ComputeMode int

const (
	COMPUTEMODE_DEFAULT ComputeMode = iota
	COMPUTEMODE_EXCLUSIVE_THREAD
	COMPUTEMODE_PROHIBITED
	COMPUTEMODE_EXCLUSIVE_PROCESS
	COMPUTEMODE_COUNT
)

type ProcessInfo struct {
	Pid           uint
	UsedGPUMemory uint64
}

type ClocksThrottleReasons uint64

const (
	ClocksThrottleReasonAll ClocksThrottleReasons = ClocksThrottleReasonApplicationsClocksSetting |
		ClocksThrottleReasonGpuIdle |
		ClocksThrottleReasonHwSlowdown |
		ClocksThrottleReasonNone |
		ClocksThrottleReasonSwPowerCap |
		ClocksThrottleReasonUnknown
	ClocksThrottleReasonApplicationsClocksSetting = 0x0000000000000002
	ClocksThrottleReasonGpuIdle                   = 0x0000000000000001
	ClocksThrottleReasonHwSlowdown                = 0x0000000000000008
	ClocksThrottleReasonNone                      = 0x0000000000000000
	ClocksThrottleReasonSwPowerCap                = 0x0000000000000004
	ClocksThrottleReasonUnknown                   = 0x8000000000000000
)

type MemoryErrorType int

const (
	MEMORY_ERROR_TYPE_CORRECTED MemoryErrorType = iota
	MEMORY_ERROR_TYPE_UNCORRECTED
	MEMORY_ERROR_TYPE_COUNT
)

type EccCounterType int

const (
	VOLATILE_ECC EccCounterType = iota
	AGGREGATE_ECC
	ECC_COUNTER_TYPE_COUNT
)

type EccErrorCounts struct {
	DeviceMemory uint64
	L1Cache      uint64
	L2Cache      uint64
	RegisterFile uint64
}

type GpuOperationMode int

const (
	GOM_ALL_ON GpuOperationMode = iota
	GOM_COMPUTE
	GOM_LOW_DP
	GOM_UNKNOWN
)

type PciInfo struct {
	BusID          string
	Domain         uint
	Bus            uint
	Device         uint
	PciDeviceID    uint
	PciSubSystemID uint
}

type InforomObject int

const (
	INFOROM_OEM InforomObject = iota
	INFOROM_ECC
	INFOROM_POWER
	INFOROM_COUNT
)

type MemoryLocation int

const (
	MEMORY_LOCATION_L1_CACHE = iota
	MEMORY_LOCATION_L2_CACHE
	MEMORY_LOCATION_DEVICE_MEMORY
	MEMORY_LOCATION_REGISTER_FILE
	MEMORY_LOCATION_TEXTURE_MEMORY
	MEMORY_LOCATION_COUNT
)

type PcieUtilCounter int

const (
	PCIE_UTIL_TX_BYTES PcieUtilCounter = iota
	PCIE_UTIL_RX_BYTES
	PCIE_UTIL_COUNT
)

type PageRetirementCause int

const (
	PAGE_RETIREMENT_CAUSE_MULTIPLE_SINGLE_BIT_ECC_ERRORS PageRetirementCause = iota
	PAGE_RETIREMENT_CAUSE_DOUBLE_BIT_ECC_ERROR
	PAGE_RETIREMENT_CAUSE_COUNT
)

type TemperatureThresholds int

const (
	TEMPERATURE_THRESHOLD_SHUTDOWN TemperatureThresholds = iota
	TEMPERATURE_THRESHOLD_SLOWDOWN
	TEMPERATURE_THRESHOLD_COUNT
)

type GpuTopologyLevel int

const (
	TOPOLOGY_INTERNAL   GpuTopologyLevel = iota
	TOPOLOGY_SINGLE                      = 10
	TOPOLOGY_MULTIPLE                    = 20
	TOPOLOGY_HOSTBRIDGE                  = 30
	TOPOLOGY_CPU                         = 40
	TOPOLOGY_SYSTEM                      = 50
	TOPOLOGY_UNKNOWN                     = 60
)

type Utilization struct {
	GPU    uint
	Memory uint
}

type PerfPolicy int

const (
	PERF_POLICY_POWER PerfPolicy = iota
	PERF_POLICY_THERMAL
	PERF_POLICY_COUNT
)

type ViolationTime struct {
	ReferTime     time.Duration
	ViolationTime time.Duration
}

type HwbcEntry struct {
	ID        uint
	FwVersion string
}

type ProcessUtilizationSample struct {
	Pid       uint
	TimeStamp time.Duration
	SmUtil    uint
	MemUtil   uint
	EncUtil   uint
	DecUtil   uint
}

type EventSet struct{ set C.nvmlEventSet_t }

type EventType uint64

const (
	EventTypeAll EventType = EventTypeNone |
		EventTypeSingleBitEccError |
		EventTypeDoubleBitEccError |
		EventTypePState |
		EventTypeXidCriticalError |
		EventTypeClock
	EventTypeNone              = 0x0000000000000000
	EventTypeSingleBitEccError = 0x0000000000000001
	EventTypeDoubleBitEccError = 0x0000000000000002
	EventTypePState            = 0x0000000000000004
	EventTypeXidCriticalError  = 0x0000000000000008
	EventTypeClock             = 0x0000000000000010
)

type EventData struct {
	Device handle
	Data   uint64
	Types  []EventType
}
