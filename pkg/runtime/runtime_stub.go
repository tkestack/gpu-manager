package runtime

import (
	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

type containerRuntimeManagerStub struct {
}

var _ ContainerRuntimeInterface = (*containerRuntimeManagerStub)(nil)

func NewContainerRuntimeManagerStub() *containerRuntimeManagerStub {
	return &containerRuntimeManagerStub{}
}

func (m *containerRuntimeManagerStub) GetPidsInContainers(containerID string) ([]int, error) {
	return nil, nil
}

func (m *containerRuntimeManagerStub) InspectContainer(containerID string) (*criapi.ContainerStatus, error) {
	return nil, nil
}

func (m *containerRuntimeManagerStub) RuntimeName() string { return "" }
