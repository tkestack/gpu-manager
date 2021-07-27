module tkestack.io/gpu-manager

go 1.14

replace tkestack.io/nvml => github.com/tkestack/go-nvml v0.0.0-20191217064248-7363e630a33e

require (
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/fsnotify/fsnotify v1.4.7
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/opencontainers/runc v1.0.0-rc95
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.2.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.27.0
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	k8s.io/cri-api v0.17.4
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.17.4
	k8s.io/kubelet v0.17.4
	tkestack.io/nvml v0.0.0-00010101000000-000000000000
)
