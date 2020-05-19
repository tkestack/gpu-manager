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

/*
#include "nvml_dl.h"
#include <stddef.h>
#include <stdlib.h>

static void *handle;

int brandType_to_int(nvmlBrandType_t t) { return (int)t; }

int bridgeChipType_to_int(nvmlBridgeChipType_t t) { return (int)t; }

int computeMode_to_int(nvmlComputeMode_t t) { return (int)t; }

int gpuOperationMode_to_int(nvmlGpuOperationMode_t t) { return (int)t; }

int pstates_to_int(nvmlPstates_t t) { return (int)t; }

int samplingType_to_int(nvmlSamplingType_t t) { return (int)t; }

int gpuTopologyLevel_to_int(nvmlGpuTopologyLevel_t t) { return (int)t; }

int perfPolicyType_to_int(nvmlPerfPolicyType_t t) { return (int)t; }

const char *NVML_DL(nvmlErrorString)(nvmlReturn_t result) {
  nvmlErrSym_t errSym;

  dlerror();
  errSym = dlsym(handle, "nvmlErrorString");
  if (dlerror()) {
    return "library not found";
  }

  return ((*errSym)(result));
}

// http://docs.nvidia.com/deploy/nvml-api/group__nvmlInitializationAndCleanup.html
nvmlReturn_t NVML_DL(nvmlInit)(void) {
  nvmlSym_t sym;

  handle = dlopen("libnvidia-ml.so", RTLD_NOW | RTLD_NODELETE | RTLD_GLOBAL);
  if (handle == NULL) {
    return (NVML_ERROR_LIBRARY_NOT_FOUND);
  }

#if NVML_API_VERSION >= 9
  DLSYM(sym, nvmlInit_v2);
#else
  DLSYM(sym, nvmlInit);
#endif

  return ((*sym)());
}

nvmlReturn_t NVML_DL(nvmlShutdown)(void) {
  nvmlSym_t sym;

  DLSYM(sym, nvmlShutdown);

  nvmlReturn_t r = ((*sym)());
  if (r != NVML_SUCCESS) {
    return (r);
  }
  return (dlclose(handle) ? NVML_ERROR_UNKNOWN : NVML_SUCCESS);
}

// http://docs.nvidia.com/deploy/nvml-api/group__nvmlSystemQueries.html
nvmlReturn_t NVML_DL(nvmlSystemGetDriverVersion)(char *version,
                                                 unsigned int length) {
  CALL(nvmlSystemGetDriverVersion, version, length);
}

nvmlReturn_t NVML_DL(nvmlSystemGetNVMLVersion)(char *version,
                                               unsigned int length) {
  CALL(nvmlSystemGetNVMLVersion, version, length);
}

nvmlReturn_t NVML_DL(nvmlSystemGetProcessName)(unsigned int pid, char *name,
                                               unsigned int length) {
  CALL(nvmlSystemGetProcessName, pid, name, length);
}

// http://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html
nvmlReturn_t NVML_DL(nvmlDeviceClearCpuAffinity)(nvmlDevice_t deviced) {
  CALL(nvmlDeviceClearCpuAffinity, deviced);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetAPIRestriction)(nvmlDevice_t device,
                                     nvmlRestrictedAPI_t apiType,
                                     nvmlEnableState_t *isRestrictedd) {
  CALL(nvmlDeviceGetAPIRestriction, device, apiType, isRestrictedd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetApplicationsClock)(nvmlDevice_t device,
                                                     nvmlClockType_t clockType,
                                                     unsigned int *clockMHzd) {
  CALL(nvmlDeviceGetApplicationsClock, device, clockType, clockMHzd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetAutoBoostedClocksEnabled)(
    nvmlDevice_t device, nvmlEnableState_t *isEnabled,
    nvmlEnableState_t *defaultIsEnabledd) {
  CALL(nvmlDeviceGetAutoBoostedClocksEnabled, device, isEnabled,
       defaultIsEnabledd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetBAR1MemoryInfo)(nvmlDevice_t device,
                                     nvmlBAR1Memory_t *bar1Memoryd) {
  CALL(nvmlDeviceGetBAR1MemoryInfo, device, bar1Memoryd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetBoardId)(nvmlDevice_t device,
                                           unsigned int *boardIdd) {
  CALL(nvmlDeviceGetBoardId, device, boardIdd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetBrand)(nvmlDevice_t device,
                                         nvmlBrandType_t *typed) {
  CALL(nvmlDeviceGetBrand, device, typed);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetBridgeChipInfo)(
    nvmlDevice_t device, nvmlBridgeChipHierarchy_t *bridgeHierarchyd) {
  CALL(nvmlDeviceGetBridgeChipInfo, device, bridgeHierarchyd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetClockInfo)(nvmlDevice_t device,
                                             nvmlClockType_t type,
                                             unsigned int *clockd) {
  CALL(nvmlDeviceGetClockInfo, device, type, clockd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetComputeMode)(nvmlDevice_t device,
                                               nvmlComputeMode_t *moded) {
  CALL(nvmlDeviceGetComputeMode, device, moded);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetComputeRunningProcesses)(
    nvmlDevice_t device, unsigned int *infoCount, nvmlProcessInfo_t *infosd) {
  CALL(nvmlDeviceGetComputeRunningProcesses, device, infoCount, infosd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetCount)(unsigned int *deviceCountd) {
#if NVML_API_VERSION >= 9
  CALL(nvmlDeviceGetCount_v2, deviceCountd);
#else
  CALL(nvmlDeviceGetCount, deviceCountd);
#endif
}

nvmlReturn_t NVML_DL(nvmlDeviceGetCpuAffinity)(nvmlDevice_t device,
                                               unsigned int cpuSetSize,
                                               unsigned long *cpuSetd) {
  CALL(nvmlDeviceGetCpuAffinity, device, cpuSetSize, cpuSetd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetCurrPcieLinkGeneration)(nvmlDevice_t device,
                                             unsigned int *currLinkGend) {
  CALL(nvmlDeviceGetCurrPcieLinkGeneration, device, currLinkGend);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetCurrPcieLinkWidth)(nvmlDevice_t device,
                                        unsigned int *currLinkWidthd) {
  CALL(nvmlDeviceGetCurrPcieLinkWidth, device, currLinkWidthd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetCurrentClocksThrottleReasons)(
    nvmlDevice_t device, unsigned long long *clocksThrottleReasonsd) {
  CALL(nvmlDeviceGetCurrentClocksThrottleReasons, device,
       clocksThrottleReasonsd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetDecoderUtilization)(nvmlDevice_t device,
                                         unsigned int *utilization,
                                         unsigned int *samplingPeriodUsd) {
  CALL(nvmlDeviceGetDecoderUtilization, device, utilization, samplingPeriodUsd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetDefaultApplicationsClock)(
    nvmlDevice_t device, nvmlClockType_t clockType, unsigned int *clockMHzd) {
  CALL(nvmlDeviceGetDefaultApplicationsClock, device, clockType, clockMHzd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetDetailedEccErrors)(
    nvmlDevice_t device, nvmlMemoryErrorType_t errorType,
    nvmlEccCounterType_t counterType, nvmlEccErrorCounts_t *eccCountsd) {
  CALL(nvmlDeviceGetDetailedEccErrors, device, errorType, counterType,
       eccCountsd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetDisplayActive)(nvmlDevice_t device,
                                                 nvmlEnableState_t *isActived) {
  CALL(nvmlDeviceGetDisplayActive, device, isActived);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetDisplayMode)(nvmlDevice_t device,
                                               nvmlEnableState_t *displayd) {
  CALL(nvmlDeviceGetDisplayMode, device, displayd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetEccMode)(nvmlDevice_t device,
                                           nvmlEnableState_t *current,
                                           nvmlEnableState_t *pendingd) {
  CALL(nvmlDeviceGetEccMode, device, current, pendingd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetEncoderUtilization)(nvmlDevice_t device,
                                         unsigned int *utilization,
                                         unsigned int *samplingPeriodUsd) {
  CALL(nvmlDeviceGetEncoderUtilization, device, utilization, samplingPeriodUsd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetEnforcedPowerLimit)(nvmlDevice_t device,
                                                      unsigned int *limitd) {
  CALL(nvmlDeviceGetEnforcedPowerLimit, device, limitd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetFanSpeed)(nvmlDevice_t device,
                                            unsigned int *speedd) {
  CALL(nvmlDeviceGetFanSpeed, device, speedd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetGpuOperationMode)(nvmlDevice_t device,
                                       nvmlGpuOperationMode_t *current,
                                       nvmlGpuOperationMode_t *pendingd) {
  CALL(nvmlDeviceGetGpuOperationMode, device, current, pendingd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetGraphicsRunningProcesses)(
    nvmlDevice_t device, unsigned int *infoCount, nvmlProcessInfo_t *infosd) {
  CALL(nvmlDeviceGetGraphicsRunningProcesses, device, infoCount, infosd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetHandleByIndex)(unsigned int index,
                                                 nvmlDevice_t *deviced) {
#if NVML_API_VERSION >= 9
  CALL(nvmlDeviceGetHandleByIndex_v2, index, deviced);
#else
  CALL(nvmlDeviceGetHandleByIndex, index, deviced);
#endif
}

nvmlReturn_t NVML_DL(nvmlDeviceGetHandleByPciBusId)(const char *pciBusId,
                                                    nvmlDevice_t *deviced) {
#if NVML_API_VERSION >= 9
  CALL(nvmlDeviceGetHandleByPciBusId_v2, pciBusId, deviced);
#else
  CALL(nvmlDeviceGetHandleByPciBusId, pciBusId, deviced);
#endif
}

nvmlReturn_t NVML_DL(nvmlDeviceGetHandleBySerial)(const char *serial,
                                                  nvmlDevice_t *deviced) {
  CALL(nvmlDeviceGetHandleBySerial, serial, deviced);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetHandleByUUID)(const char *uuid,
                                                nvmlDevice_t *deviced) {
  CALL(nvmlDeviceGetHandleByUUID, uuid, deviced);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetIndex)(nvmlDevice_t device,
                                         unsigned int *indexd) {
  CALL(nvmlDeviceGetIndex, device, indexd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetInforomConfigurationChecksum)(nvmlDevice_t device,
                                                   unsigned int *checksumd) {
  CALL(nvmlDeviceGetInforomConfigurationChecksum, device, checksumd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetInforomImageVersion)(nvmlDevice_t device,
                                                       char *version,
                                                       unsigned int lengthd) {
  CALL(nvmlDeviceGetInforomImageVersion, device, version, lengthd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetInforomVersion)(nvmlDevice_t device,
                                                  nvmlInforomObject_t object,
                                                  char *version,
                                                  unsigned int lengthd) {
  CALL(nvmlDeviceGetInforomVersion, device, object, version, lengthd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetMaxClockInfo)(nvmlDevice_t device,
                                                nvmlClockType_t type,
                                                unsigned int *clockd) {
  CALL(nvmlDeviceGetMaxClockInfo, device, type, clockd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetMaxPcieLinkGeneration)(nvmlDevice_t device,
                                            unsigned int *maxLinkGend) {
  CALL(nvmlDeviceGetMaxPcieLinkGeneration, device, maxLinkGend);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetMaxPcieLinkWidth)(nvmlDevice_t device,
                                       unsigned int *maxLinkWidthd) {
  CALL(nvmlDeviceGetMaxPcieLinkWidth, device, maxLinkWidthd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetMemoryErrorCounter)(
    nvmlDevice_t device, nvmlMemoryErrorType_t errorType,
    nvmlEccCounterType_t counterType, nvmlMemoryLocation_t locationType,
    unsigned long long *countd) {
  CALL(nvmlDeviceGetMemoryErrorCounter, device, errorType, counterType,
       locationType, countd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetMemoryInfo)(nvmlDevice_t device,
                                              nvmlMemory_t *memoryd) {
  CALL(nvmlDeviceGetMemoryInfo, device, memoryd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetMinorNumber)(nvmlDevice_t device,
                                               unsigned int *minorNumberd) {
  CALL(nvmlDeviceGetMinorNumber, device, minorNumberd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetMultiGpuBoard)(nvmlDevice_t device,
                                                 unsigned int *multiGpuBoold) {
  CALL(nvmlDeviceGetMultiGpuBoard, device, multiGpuBoold);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetName)(nvmlDevice_t device, char *name,
                                        unsigned int lengthd) {
  CALL(nvmlDeviceGetName, device, name, lengthd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetPciInfo)(nvmlDevice_t device,
                                           nvmlPciInfo_t *pcid) {
#if NVML_API_VERSION >= 9
  CALL(nvmlDeviceGetPciInfo_v3, device, pcid);
#else
  CALL(nvmlDeviceGetPciInfo, device, pcid);
#endif
}

nvmlReturn_t NVML_DL(nvmlDeviceGetPcieReplayCounter)(nvmlDevice_t device,
                                                     unsigned int *valued) {
  CALL(nvmlDeviceGetPcieReplayCounter, device, valued);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetPcieThroughput)(nvmlDevice_t device,
                                                  nvmlPcieUtilCounter_t counter,
                                                  unsigned int *valued) {
  CALL(nvmlDeviceGetPcieThroughput, device, counter, valued);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetPerformanceState)(nvmlDevice_t device,
                                                    nvmlPstates_t *pStated) {
  CALL(nvmlDeviceGetPerformanceState, device, pStated);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetPersistenceMode)(nvmlDevice_t device,
                                                   nvmlEnableState_t *moded) {
  CALL(nvmlDeviceGetPersistenceMode, device, moded);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetPowerManagementDefaultLimit)(nvmlDevice_t device,
                                                  unsigned int *defaultLimitd) {
  CALL(nvmlDeviceGetPowerManagementDefaultLimit, device, defaultLimitd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetPowerManagementLimit)(nvmlDevice_t device,
                                                        unsigned int *limitd) {
  CALL(nvmlDeviceGetPowerManagementLimit, device, limitd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetPowerManagementLimitConstraints)(
    nvmlDevice_t device, unsigned int *minLimit, unsigned int *maxLimitd) {
  CALL(nvmlDeviceGetPowerManagementLimitConstraints, device, minLimit,
       maxLimitd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetPowerManagementMode)(nvmlDevice_t device,
                                          nvmlEnableState_t *moded) {
  CALL(nvmlDeviceGetPowerManagementMode, device, moded);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetPowerState)(nvmlDevice_t device,
                                              nvmlPstates_t *pStated) {
  CALL(nvmlDeviceGetPowerState, device, pStated);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetPowerUsage)(nvmlDevice_t device,
                                              unsigned int *powerd) {
  CALL(nvmlDeviceGetPowerUsage, device, powerd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetRetiredPages)(
    nvmlDevice_t device, nvmlPageRetirementCause_t cause,
    unsigned int *pageCount, unsigned long long *addressesd) {
  CALL(nvmlDeviceGetRetiredPages, device, cause, pageCount, addressesd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetRetiredPagesPendingStatus)(nvmlDevice_t device,
                                                nvmlEnableState_t *isPendingd) {
  CALL(nvmlDeviceGetRetiredPagesPendingStatus, device, isPendingd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetSamples)(nvmlDevice_t device,
                                           nvmlSamplingType_t type,
                                           unsigned long long lastSeenTimeStamp,
                                           nvmlValueType_t *sampleValType,
                                           unsigned int *sampleCount,
                                           nvmlSample_t *samplesd) {
  CALL(nvmlDeviceGetSamples, device, type, lastSeenTimeStamp, sampleValType,
       sampleCount, samplesd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetSerial)(nvmlDevice_t device, char *serial,
                                          unsigned int lengthd) {
  CALL(nvmlDeviceGetSerial, device, serial, lengthd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetSupportedClocksThrottleReasons)(
    nvmlDevice_t device, unsigned long long *supportedClocksThrottleReasonsd) {
  CALL(nvmlDeviceGetSupportedClocksThrottleReasons, device,
       supportedClocksThrottleReasonsd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetSupportedGraphicsClocks)(
    nvmlDevice_t device, unsigned int memoryClockMHz, unsigned int *count,
    unsigned int *clocksMHzd) {
  CALL(nvmlDeviceGetSupportedGraphicsClocks, device, memoryClockMHz, count,
       clocksMHzd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetSupportedMemoryClocks)(
    nvmlDevice_t device, unsigned int *count, unsigned int *clocksMHzd) {
  CALL(nvmlDeviceGetSupportedMemoryClocks, device, count, clocksMHzd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetTemperature)(nvmlDevice_t device,
                                  nvmlTemperatureSensors_t sensorType,
                                  unsigned int *tempd) {
  CALL(nvmlDeviceGetTemperature, device, sensorType, tempd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetTemperatureThreshold)(
    nvmlDevice_t device, nvmlTemperatureThresholds_t thresholdType,
    unsigned int *tempd) {
  CALL(nvmlDeviceGetTemperatureThreshold, device, thresholdType, tempd);
}
nvmlReturn_t NVML_DL(nvmlDeviceGetTopologyCommonAncestor)(
    nvmlDevice_t device1, nvmlDevice_t device2,
    nvmlGpuTopologyLevel_t *pathInfod) {
  CALL(nvmlDeviceGetTopologyCommonAncestor, device1, device2, pathInfod);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetTopologyNearestGpus)(
    nvmlDevice_t device, nvmlGpuTopologyLevel_t level, unsigned int *count,
    nvmlDevice_t *deviceArrayd) {
  CALL(nvmlDeviceGetTopologyNearestGpus, device, level, count, deviceArrayd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetTotalEccErrors)(
    nvmlDevice_t device, nvmlMemoryErrorType_t errorType,
    nvmlEccCounterType_t counterType, unsigned long long *eccCountsd) {
  CALL(nvmlDeviceGetTotalEccErrors, device, errorType, counterType, eccCountsd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetUUID)(nvmlDevice_t device, char *uuid,
                                        unsigned int lengthd) {
  CALL(nvmlDeviceGetUUID, device, uuid, lengthd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetUtilizationRates)(nvmlDevice_t device,
                                       nvmlUtilization_t *utilizationd) {
  CALL(nvmlDeviceGetUtilizationRates, device, utilizationd);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetVbiosVersion)(nvmlDevice_t device,
                                                char *version,
                                                unsigned int lengthd) {
  CALL(nvmlDeviceGetVbiosVersion, device, version, lengthd);
}

nvmlReturn_t
NVML_DL(nvmlDeviceGetViolationStatus)(nvmlDevice_t device,
                                      nvmlPerfPolicyType_t perfPolicyType,
                                      nvmlViolationTime_t *violTimed) {
  CALL(nvmlDeviceGetViolationStatus, device, perfPolicyType, violTimed);
}

nvmlReturn_t NVML_DL(nvmlDeviceOnSameBoard)(nvmlDevice_t device1,
                                            nvmlDevice_t device2,
                                            int *onSameBoardd) {
  CALL(nvmlDeviceOnSameBoard, device1, device2, onSameBoardd);
}

nvmlReturn_t NVML_DL(nvmlDeviceResetApplicationsClocks)(nvmlDevice_t deviced) {
  CALL(nvmlDeviceResetApplicationsClocks, deviced);
}

nvmlReturn_t
NVML_DL(nvmlDeviceSetAutoBoostedClocksEnabled)(nvmlDevice_t device,
                                               nvmlEnableState_t enabledd) {
  CALL(nvmlDeviceSetAutoBoostedClocksEnabled, device, enabledd);
}

nvmlReturn_t NVML_DL(nvmlDeviceSetCpuAffinity)(nvmlDevice_t deviced) {
  CALL(nvmlDeviceSetCpuAffinity, deviced);
}

nvmlReturn_t NVML_DL(nvmlDeviceSetDefaultAutoBoostedClocksEnabled)(
    nvmlDevice_t device, nvmlEnableState_t enabled, unsigned int flagsd) {
  CALL(nvmlDeviceSetDefaultAutoBoostedClocksEnabled, device, enabled, flagsd);
}

nvmlReturn_t NVML_DL(nvmlDeviceValidateInforom)(nvmlDevice_t deviced) {
  CALL(nvmlDeviceValidateInforom, deviced);
}

nvmlReturn_t NVML_DL(nvmlSystemGetTopologyGpuSet)(unsigned int cpuNumber,
                                                  unsigned int *count,
                                                  nvmlDevice_t *deviceArrayd) {
  CALL(nvmlSystemGetTopologyGpuSet, cpuNumber, count, deviceArrayd);
}

// http://docs.nvidia.com/deploy/nvml-api/group__nvmlUnitQueries.html
nvmlReturn_t NVML_DL(nvmlSystemGetHicVersion)(unsigned int *hwbcCount,
                                              nvmlHwbcEntry_t *hwbcEntries) {
  CALL(nvmlSystemGetHicVersion, hwbcCount, hwbcEntries);
}

// http://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceCommands.html
nvmlReturn_t
NVML_DL(nvmlDeviceClearEccErrorCounts)(nvmlDevice_t device,
                                       nvmlEccCounterType_t counterType) {
  CALL(nvmlDeviceClearEccErrorCounts, device, counterType);
}
nvmlReturn_t
NVML_DL(nvmlDeviceSetAPIRestriction)(nvmlDevice_t device,
                                     nvmlRestrictedAPI_t apiType,
                                     nvmlEnableState_t isRestricted) {
  CALL(nvmlDeviceSetAPIRestriction, device, apiType, isRestricted);
}

nvmlReturn_t
NVML_DL(nvmlDeviceSetApplicationsClocks)(nvmlDevice_t device,
                                         unsigned int memClockMHz,
                                         unsigned int graphicsClockMHz) {
  CALL(nvmlDeviceSetApplicationsClocks, device, memClockMHz, graphicsClockMHz);
}

nvmlReturn_t NVML_DL(nvmlDeviceSetComputeMode)(nvmlDevice_t device,
                                               nvmlComputeMode_t mode) {
  CALL(nvmlDeviceSetComputeMode, device, mode);
}

nvmlReturn_t NVML_DL(nvmlDeviceSetEccMode)(nvmlDevice_t device,
                                           nvmlEnableState_t ecc) {
  CALL(nvmlDeviceSetEccMode, device, ecc);
}

nvmlReturn_t
NVML_DL(nvmlDeviceSetGpuOperationMode)(nvmlDevice_t device,
                                       nvmlGpuOperationMode_t mode) {
  CALL(nvmlDeviceSetGpuOperationMode, device, mode);
}

nvmlReturn_t NVML_DL(nvmlDeviceSetPersistenceMode)(nvmlDevice_t device,
                                                   nvmlEnableState_t mode) {
  CALL(nvmlDeviceSetPersistenceMode, device, mode);
}

nvmlReturn_t NVML_DL(nvmlDeviceSetPowerManagementLimit)(nvmlDevice_t device,
                                                        unsigned int limit) {
  CALL(nvmlDeviceSetPowerManagementLimit, device, limit);
}

nvmlReturn_t NVML_DL(nvmlDeviceGetProcessUtilization)(
    nvmlDevice_t device, nvmlProcessUtilizationSample_t *utilization,
    unsigned int *processSamplesCount, unsigned long long lastSeenTimeStamp) {
  CALL(nvmlDeviceGetProcessUtilization, device, utilization, processSamplesCount, lastSeenTimeStamp);
}

nvmlReturn_t __nvmlDeviceGetAverageUsage(nvmlDevice_t device,
                                         nvmlSamplingType_t type,
                                         unsigned long long lastSeenTimeStamp,
                                         unsigned int *averageUsage) {
  nvmlValueType_t sampleValType;
  unsigned int sampleCount = 0;
  nvmlSym_t hdl = NULL;
  nvmlSample_t *samples = NULL;
  int i = 0;
  unsigned int sum = 0;

  DLSYM(hdl, nvmlDeviceGetSamples);
  nvmlReturn_t r =
      hdl(device, type, lastSeenTimeStamp, &sampleValType, &sampleCount, NULL);
  if (r != NVML_SUCCESS) {
    return r;
  }

  samples = (nvmlSample_t *)malloc(sampleCount * sizeof(nvmlSample_t));

  r = hdl(device, type, lastSeenTimeStamp, &sampleValType, &sampleCount,
          samples);
  if (r != NVML_SUCCESS) {
    free(samples);
    return r;
  }

  for (; i < sampleCount; i++) {
    sum += samples[i].sampleValue.uiVal;
  }
  *averageUsage = sum / sampleCount;

  free(samples);

  return r;
}


nvmlReturn_t NVML_DL(nvmlDeviceGetSupportedEventTypes)(nvmlDevice_t device,
                                              unsigned long long *eventTypes) {
  CALL(nvmlDeviceGetSupportedEventTypes, device, eventTypes);
}

nvmlReturn_t NVML_DL(nvmlDeviceRegisterEvents)(nvmlDevice_t device,
                                             unsigned long long eventTypes,
                                             nvmlEventSet_t set) {
  CALL(nvmlDeviceRegisterEvents, device, eventTypes, set);
}

nvmlReturn_t NVML_DL(nvmlEventSetCreate)(nvmlEventSet_t *set) {
  CALL(nvmlEventSetCreate, set);
}

nvmlReturn_t NVML_DL(nvmlEventSetFree)(nvmlEventSet_t set) {
  CALL(nvmlEventSetFree, set);
}

nvmlReturn_t NVML_DL(nvmlEventSetWait)(nvmlEventSet_t set, nvmlEventData_t *data,
                                     unsigned int timeoutms) {
  CALL(nvmlEventSetWait, set, data, timeoutms);
}
*/
// #cgo CFLAGS: -I. -I /usr/local/cuda/include
// #cgo LDFLAGS: -ldl -Wl,--unresolved-symbols=ignore-in-object-files
import "C"
