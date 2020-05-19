package docker

import (
	"context"

	dockerapi "github.com/docker/docker/client"
	"k8s.io/klog"

	"tkestack.io/gpu-manager/pkg/runtime"
)

type dockerRuntimeManager struct {
	client *dockerapi.Client
}

var _ runtime.ContainerRuntimeInterface = (*dockerRuntimeManager)(nil)

func NewDockerRuntimeManager(endpoint string) *dockerRuntimeManager {
	cli, err := dockerapi.NewClientWithOpts(dockerapi.WithHost(endpoint), dockerapi.WithVersion(""))
	if err != nil {
		klog.Exitf("can't create docker client: %v", err)
		return nil
	}

	m := &dockerRuntimeManager{
		client: cli,
	}

	return m
}

func (m *dockerRuntimeManager) CgroupDriver() (string, error) {
	info, err := m.client.Info(context.Background())
	if err != nil {
		klog.Errorf("can't get cgroup driver name: %v", err)
		return "", err
	}

	return info.CgroupDriver, nil
}

func (m *dockerRuntimeManager) InspectContainer(name string) (*runtime.ContainerInfo, error) {
	jsonData, err := m.client.ContainerInspect(context.Background(), name)
	if err != nil {
		klog.Errorf("can't get container %s information: %v", name, err)
		return nil, err
	}

	return &runtime.ContainerInfo{
		Name:         jsonData.Name,
		ID:           jsonData.ID,
		Labels:       jsonData.Config.Labels,
		CgroupParent: jsonData.HostConfig.CgroupParent,
	}, nil
}
