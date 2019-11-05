# GPU Manager

GPU Manager is used for managing the nvidia GPU devices in Kubernetes cluster. It implements the `DevicePlugin` interface
of Kubernetes. So it's compatible with 1.9+ of Kubernetes release version. 

To compare with the combination solution of `nvidia-docker`
and `nvidia-k8s-plugin`, GPU manager will use native `runc` without modification but nvidia solution does.
Besides we also support metrics report without deploying new components. 

To schedule a GPU payload correctly, GPU manager should work with `gpu-quota-admission` which is a kubernetes scheduler plugin.

GPU manager also supports the payload with fraction resource of GPU device such as 0.1 card or 100MiB gpu device memory.
If you want this kind feature, please refer to `vcuda` project.

# How to deploy GPU Manager

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

- change gpu-manager.yaml and submit

change --incluster-mode from `false` to `true`, change image field to `<your repository>/public/gpu-manager:latest`, add serviceAccount filed to `gpu-manager-role`
