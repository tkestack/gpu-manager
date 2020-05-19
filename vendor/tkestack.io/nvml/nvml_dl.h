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

#ifndef NVML_DL_H
#define NVML_DL_H

#include <dlfcn.h>
#include <stddef.h>

#include "nvml.h"

#define NVML_DL(x) x##_dlib

#define DLSYM(x, sym)                                                          \
  do {                                                                         \
    dlerror();                                                                 \
    x = dlsym(handle, #sym);                                                   \
    if (dlerror() != NULL) {                                                   \
      return (NVML_ERROR_FUNCTION_NOT_FOUND);                                  \
    }                                                                          \
  } while (0)

#define CALL(func, ...)                                                        \
  do {                                                                         \
    nvmlSym_t hdl;                                                             \
    DLSYM(hdl, func);                                                          \
    return ((*hdl))(__VA_ARGS__);                                              \
  } while (0)

typedef nvmlReturn_t (*nvmlSym_t)();
typedef const char *(*nvmlErrSym_t)(nvmlReturn_t result);

extern int brandType_to_int(nvmlBrandType_t t);
extern int bridgeChipType_to_int(nvmlBridgeChipType_t t);
extern int computeMode_to_int(nvmlComputeMode_t t);
extern int gpuOperationMode_to_int(nvmlGpuOperationMode_t t);
extern int pstates_to_int(nvmlPstates_t t);
extern int samplingType_to_int(nvmlSamplingType_t t);
extern int gpuTopologyLevel_to_int(nvmlGpuTopologyLevel_t t);
extern int perfPolicyType_to_int(nvmlPerfPolicyType_t t);

extern const char *NVML_DL(nvmlErrorString)(nvmlReturn_t result);

// http://docs.nvidia.com/deploy/nvml-api/group__nvmlInitializationAndCleanup.html
extern nvmlReturn_t NVML_DL(nvmlInit)(void);
extern nvmlReturn_t NVML_DL(nvmlShutdown)(void);

// http://docs.nvidia.com/deploy/nvml-api/group__nvmlSystemQueries.html
extern nvmlReturn_t NVML_DL(nvmlSystemGetDriverVersion)(char *version,
                                                        unsigned int length);
extern nvmlReturn_t NVML_DL(nvmlSystemGetNVMLVersion)(char *version,
                                                      unsigned int length);
extern nvmlReturn_t NVML_DL(nvmlSystemGetProcessName)(unsigned int pid,
                                                      char *name,
                                                      unsigned int length);

// http://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceQueries.html
extern nvmlReturn_t NVML_DL(nvmlDeviceClearCpuAffinity)(nvmlDevice_t deviced);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetAPIRestriction)(nvmlDevice_t device,
                                         nvmlRestrictedAPI_t apiType,
                                         nvmlEnableState_t *isRestrictedd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetApplicationsClock)(
    nvmlDevice_t device, nvmlClockType_t clockType, unsigned int *clockMHzd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetAutoBoostedClocksEnabled)(
    nvmlDevice_t device, nvmlEnableState_t *isEnabled,
    nvmlEnableState_t *defaultIsEnabledd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetBAR1MemoryInfo)(nvmlDevice_t device,
                                         nvmlBAR1Memory_t *bar1Memoryd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetBoardId)(nvmlDevice_t device,
                                                  unsigned int *boardIdd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetBrand)(nvmlDevice_t device,
                                                nvmlBrandType_t *typed);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetBridgeChipInfo)(
    nvmlDevice_t device, nvmlBridgeChipHierarchy_t *bridgeHierarchyd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetClockInfo)(nvmlDevice_t device,
                                                    nvmlClockType_t type,
                                                    unsigned int *clockd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetComputeMode)(nvmlDevice_t device,
                                                      nvmlComputeMode_t *moded);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetComputeRunningProcesses)(
    nvmlDevice_t device, unsigned int *infoCount, nvmlProcessInfo_t *infosd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetCount)(unsigned int *deviceCountd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetCpuAffinity)(nvmlDevice_t device,
                                                      unsigned int cpuSetSize,
                                                      unsigned long *cpuSetd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetCurrPcieLinkGeneration)(nvmlDevice_t device,
                                                 unsigned int *currLinkGend);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetCurrPcieLinkWidth)(nvmlDevice_t device,
                                            unsigned int *currLinkWidthd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetCurrentClocksThrottleReasons)(
    nvmlDevice_t device, unsigned long long *clocksThrottleReasonsd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetDecoderUtilization)(nvmlDevice_t device,
                                             unsigned int *utilization,
                                             unsigned int *samplingPeriodUsd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetDefaultApplicationsClock)(
    nvmlDevice_t device, nvmlClockType_t clockType, unsigned int *clockMHzd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetDetailedEccErrors)(
    nvmlDevice_t device, nvmlMemoryErrorType_t errorType,
    nvmlEccCounterType_t counterType, nvmlEccErrorCounts_t *eccCountsd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetDisplayActive)(nvmlDevice_t device,
                                        nvmlEnableState_t *isActived);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetDisplayMode)(nvmlDevice_t device,
                                      nvmlEnableState_t *displayd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetEccMode)(nvmlDevice_t device,
                                                  nvmlEnableState_t *current,
                                                  nvmlEnableState_t *pendingd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetEncoderUtilization)(nvmlDevice_t device,
                                             unsigned int *utilization,
                                             unsigned int *samplingPeriodUsd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetEnforcedPowerLimit)(nvmlDevice_t device,
                                             unsigned int *limitd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetFanSpeed)(nvmlDevice_t device,
                                                   unsigned int *speedd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetGpuOperationMode)(nvmlDevice_t device,
                                           nvmlGpuOperationMode_t *current,
                                           nvmlGpuOperationMode_t *pendingd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetGraphicsRunningProcesses)(
    nvmlDevice_t device, unsigned int *infoCount, nvmlProcessInfo_t *infosd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetHandleByIndex)(unsigned int index,
                                                        nvmlDevice_t *deviced);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetHandleByPciBusId)(const char *pciBusId,
                                           nvmlDevice_t *deviced);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetHandleBySerial)(const char *serial,
                                                         nvmlDevice_t *deviced);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetHandleByUUID)(const char *uuid,
                                                       nvmlDevice_t *deviced);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetIndex)(nvmlDevice_t device,
                                                unsigned int *indexd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetInforomConfigurationChecksum)(nvmlDevice_t device,
                                                       unsigned int *checksumd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetInforomImageVersion)(
    nvmlDevice_t device, char *version, unsigned int lengthd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetInforomVersion)(nvmlDevice_t device,
                                         nvmlInforomObject_t object,
                                         char *version, unsigned int lengthd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetMaxClockInfo)(nvmlDevice_t device,
                                                       nvmlClockType_t type,
                                                       unsigned int *clockd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetMaxPcieLinkGeneration)(nvmlDevice_t device,
                                                unsigned int *maxLinkGend);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetMaxPcieLinkWidth)(nvmlDevice_t device,
                                           unsigned int *maxLinkWidthd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetMemoryErrorCounter)(
    nvmlDevice_t device, nvmlMemoryErrorType_t errorType,
    nvmlEccCounterType_t counterType, nvmlMemoryLocation_t locationType,
    unsigned long long *countd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetMemoryInfo)(nvmlDevice_t device,
                                                     nvmlMemory_t *memoryd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetMinorNumber)(nvmlDevice_t device,
                                      unsigned int *minorNumberd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetMultiGpuBoard)(nvmlDevice_t device,
                                        unsigned int *multiGpuBoold);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetName)(nvmlDevice_t device, char *name,
                                               unsigned int lengthd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetPciInfo)(nvmlDevice_t device,
                                                  nvmlPciInfo_t *pcid);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetPcieReplayCounter)(nvmlDevice_t device,
                                            unsigned int *valued);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetPcieThroughput)(
    nvmlDevice_t device, nvmlPcieUtilCounter_t counter, unsigned int *valued);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetPerformanceState)(nvmlDevice_t device,
                                           nvmlPstates_t *pStated);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetPersistenceMode)(nvmlDevice_t device,
                                          nvmlEnableState_t *moded);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetPowerManagementDefaultLimit)(
    nvmlDevice_t device, unsigned int *defaultLimitd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetPowerManagementLimit)(nvmlDevice_t device,
                                               unsigned int *limitd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetPowerManagementLimitConstraints)(
    nvmlDevice_t device, unsigned int *minLimit, unsigned int *maxLimitd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetPowerManagementMode)(nvmlDevice_t device,
                                              nvmlEnableState_t *moded);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetPowerState)(nvmlDevice_t device,
                                                     nvmlPstates_t *pStated);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetPowerUsage)(nvmlDevice_t device,
                                                     unsigned int *powerd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetRetiredPages)(
    nvmlDevice_t device, nvmlPageRetirementCause_t cause,
    unsigned int *pageCount, unsigned long long *addressesd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetRetiredPagesPendingStatus)(
    nvmlDevice_t device, nvmlEnableState_t *isPendingd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetSamples)(
    nvmlDevice_t device, nvmlSamplingType_t type,
    unsigned long long lastSeenTimeStamp, nvmlValueType_t *sampleValType,
    unsigned int *sampleCount, nvmlSample_t *samplesd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetSerial)(nvmlDevice_t device,
                                                 char *serial,
                                                 unsigned int lengthd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetSupportedClocksThrottleReasons)(
    nvmlDevice_t device, unsigned long long *supportedClocksThrottleReasonsd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetSupportedGraphicsClocks)(
    nvmlDevice_t device, unsigned int memoryClockMHz, unsigned int *count,
    unsigned int *clocksMHzd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetSupportedMemoryClocks)(
    nvmlDevice_t device, unsigned int *count, unsigned int *clocksMHzd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetTemperature)(nvmlDevice_t device,
                                      nvmlTemperatureSensors_t sensorType,
                                      unsigned int *tempd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetTemperatureThreshold)(
    nvmlDevice_t device, nvmlTemperatureThresholds_t thresholdType,
    unsigned int *tempd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetTopologyCommonAncestor)(
    nvmlDevice_t device1, nvmlDevice_t device2,
    nvmlGpuTopologyLevel_t *pathInfod);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetTopologyNearestGpus)(
    nvmlDevice_t device, nvmlGpuTopologyLevel_t level, unsigned int *count,
    nvmlDevice_t *deviceArrayd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetTotalEccErrors)(
    nvmlDevice_t device, nvmlMemoryErrorType_t errorType,
    nvmlEccCounterType_t counterType, unsigned long long *eccCountsd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetUUID)(nvmlDevice_t device, char *uuid,
                                               unsigned int lengthd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetUtilizationRates)(nvmlDevice_t device,
                                           nvmlUtilization_t *utilizationd);
extern nvmlReturn_t NVML_DL(nvmlDeviceGetVbiosVersion)(nvmlDevice_t device,
                                                       char *version,
                                                       unsigned int lengthd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetViolationStatus)(nvmlDevice_t device,
                                          nvmlPerfPolicyType_t perfPolicyType,
                                          nvmlViolationTime_t *violTimed);
extern nvmlReturn_t NVML_DL(nvmlDeviceOnSameBoard)(nvmlDevice_t device1,
                                                   nvmlDevice_t device2,
                                                   int *onSameBoardd);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceResetApplicationsClocks)(nvmlDevice_t deviced);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceSetAutoBoostedClocksEnabled)(nvmlDevice_t device,
                                                   nvmlEnableState_t enabledd);
extern nvmlReturn_t NVML_DL(nvmlDeviceSetCpuAffinity)(nvmlDevice_t deviced);
extern nvmlReturn_t NVML_DL(nvmlDeviceSetDefaultAutoBoostedClocksEnabled)(
    nvmlDevice_t device, nvmlEnableState_t enabled, unsigned int flagsd);
extern nvmlReturn_t NVML_DL(nvmlDeviceValidateInforom)(nvmlDevice_t deviced);
extern nvmlReturn_t NVML_DL(nvmlSystemGetTopologyGpuSet)(
    unsigned int cpuNumber, unsigned int *count, nvmlDevice_t *deviceArrayd);

extern nvmlReturn_t NVML_DL(nvmlDeviceGetTopologyCommonAncestor)(
    nvmlDevice_t dev1, nvmlDevice_t dev2, nvmlGpuTopologyLevel_t *info);

// http://docs.nvidia.com/deploy/nvml-api/group__nvmlUnitQueries.html
extern nvmlReturn_t NVML_DL(nvmlSystemGetHicVersion)(
    unsigned int *hwbcCount,
    nvmlHwbcEntry_t *
        hwbcEntries); // http://docs.nvidia.com/deploy/nvml-api/group__nvmlDeviceCommands.html
extern nvmlReturn_t
    NVML_DL(nvmlDeviceClearEccErrorCounts)(nvmlDevice_t device,
                                           nvmlEccCounterType_t counterType);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceSetAPIRestriction)(nvmlDevice_t device,
                                         nvmlRestrictedAPI_t apiType,
                                         nvmlEnableState_t isRestricted);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceSetApplicationsClocks)(nvmlDevice_t device,
                                             unsigned int memClockMHz,
                                             unsigned int graphicsClockMHz);
extern nvmlReturn_t NVML_DL(nvmlDeviceSetComputeMode)(nvmlDevice_t device,
                                                      nvmlComputeMode_t mode);
extern nvmlReturn_t NVML_DL(nvmlDeviceSetEccMode)(nvmlDevice_t device,
                                                  nvmlEnableState_t ecc);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceSetGpuOperationMode)(nvmlDevice_t device,
                                           nvmlGpuOperationMode_t mode);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceSetPersistenceMode)(nvmlDevice_t device,
                                          nvmlEnableState_t mode);
extern nvmlReturn_t
    NVML_DL(nvmlDeviceSetPowerManagementLimit)(nvmlDevice_t device,
                                               unsigned int limit);

nvmlReturn_t __nvmlDeviceGetAverageUsage(nvmlDevice_t device,
                                         nvmlSamplingType_t type,
                                         unsigned long long lastSeenTimeStamp,
                                         unsigned int *averageUsage);

extern nvmlReturn_t NVML_DL(nvmlDeviceGetProcessUtilization)(
    nvmlDevice_t device, nvmlProcessUtilizationSample_t *utilization,
    unsigned int *processSamplesCount, unsigned long long lastSeenTimeStamp);

extern nvmlReturn_t
    NVML_DL(nvmlDeviceGetSupportedEventTypes)(nvmlDevice_t device,
                                              unsigned long long *eventTypes);
extern nvmlReturn_t NVML_DL(nvmlDeviceRegisterEvents)(
    nvmlDevice_t device, unsigned long long eventTypes, nvmlEventSet_t set);
extern nvmlReturn_t NVML_DL(nvmlEventSetCreate)(nvmlEventSet_t *set);
extern nvmlReturn_t NVML_DL(nvmlEventSetFree)(nvmlEventSet_t set);
extern nvmlReturn_t NVML_DL(nvmlEventSetWait)(nvmlEventSet_t set,
                                              nvmlEventData_t *data,
                                              unsigned int timeoutms);

#endif
