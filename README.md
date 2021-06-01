# GPU Manager

[![Build Status](https://travis-ci.org/tkestack/gpu-manager.svg?branch=master)](https://travis-ci.org/tkestack/gpu-manager)

GPU Manager is used for managing the nvidia GPU devices in Kubernetes cluster. It implements the `DevicePlugin` interface
of Kubernetes. So it's compatible with 1.9+ of Kubernetes release version. 

To compare with the combination solution of `nvidia-docker`
and `nvidia-k8s-plugin`, GPU manager will use native `runc` without modification but nvidia solution does.
Besides we also support metrics report without deploying new components. 

To schedule a GPU payload correctly, GPU manager should work with [gpu-admission](https://github.com/tkestack/gpu-admission) which is a
 kubernetes scheduler plugin.

GPU manager also supports the payload with fraction resource of GPU device such as 0.1 card or 100MiB gpu device memory.
If you want this kind feature, please refer to [vcuda-controller](https://github.com/tkestack/vcuda-controller) project.

## Build

**1.** Build binary

- Prerequisite
   - CUDA toolkit
    
```
make
```

**2.** Build image

- Prerequisite
    - Docker

```
make img
```

## Prebuilt image

Prebuilt image can be found at `thomassong/gpu-manager`

## Deploy

GPU Manager is running as daemonset, and because of the RABC restriction and hydrid cluster,
you need to do the following steps to make this daemonset run correctly.

- service account and clusterrole

```
kubectl create sa gpu-manager -n kube-system
kubectl create clusterrolebinding gpu-manager-role --clusterrole=cluster-admin --serviceaccount=kube-system:gpu-manager
```

- label node with `nvidia-device-enable=enable`

```
kubectl label node <node> nvidia-device-enable=enable
```

- submit daemonset yaml

```
kubectl create -f gpu-manager.yaml
```

## Pod template example

There is nothing special to submit a Pod except the description of GPU resource is no longer 1
. The GPU
resources are described as that 100 `tencent.com/vcuda-core` for 1 GPU and N `tencent.com/vcuda-memory` for GPU memory (1 tencent.com/vcuda-memory means 256Mi
GPU memory). And because of the limitation of extend resource validation of Kubernetes, to support
GPU utilization limitation, you should add `tencent.com/vcuda-core-limit: XX` in the annotation
 field of a Pod.
 
 **Notice: the value of `tencent.com/vcuda-core` is either the multiple of 100 or any value
smaller than 100.For example, 100, 200 or 20 is valid value but 150 or 250 is invalid**

- Submit a Pod with 0.3 GPU utilization and 7680MiB GPU memory with 0.5 GPU utilization limit

```
apiVersion: v1
kind: Pod
metadata:
  name: vcuda
  annotations:
    tencent.com/vcuda-core-limit: 50
spec:
  restartPolicy: Never
  containers:
  - image: <test-image>
    name: nvidia
    command:
    - /usr/local/nvidia/bin/nvidia-smi
    - pmon
    - -d
    - 10
    resources:
      requests:
        tencent.com/vcuda-core: 50
        tencent.com/vcuda-memory: 30
      limits:
        tencent.com/vcuda-core: 50
        tencent.com/vcuda-memory: 30
```

- Submit a Pod with 2 GPU card

```
apiVersion: v1
kind: Pod
metadata:
  name: vcuda
spec:
  restartPolicy: Never
  containers:
  - image: <test-image>
    name: nvidia
    command:
    - /usr/local/nvidia/bin/nvidia-smi
    - pmon
    - -d
    - 10
    resources:
      requests:
        tencent.com/vcuda-core: 200
        tencent.com/vcuda-memory: 60
      limits:
        tencent.com/vcuda-core: 200
        tencent.com/vcuda-memory: 60
```

## FAQ

If you have some questions about this project, you can first refer to [FAQ](./docs/faq.md) to find a solution.
