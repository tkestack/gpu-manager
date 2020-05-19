package fake

import (
	"fmt"

	"tkestack.io/gpu-manager/pkg/runtime"
)

type fakeRuntimeManager struct {
	Containers map[string]*runtime.ContainerInfo
}

var _ runtime.ContainerRuntimeInterface = (*fakeRuntimeManager)(nil)

func NewFakeRuntimeManager() (*fakeRuntimeManager, error) {
	return &fakeRuntimeManager{
		Containers: make(map[string]*runtime.ContainerInfo, 0),
	}, nil
}

func (f *fakeRuntimeManager) CgroupDriver() (string, error) {
	return "fake", nil
}

func (f *fakeRuntimeManager) InspectContainer(name string) (*runtime.ContainerInfo, error) {
	c, found := f.Containers[name]
	if !found {
		return nil, fmt.Errorf("container not found")
	}

	return c, nil
}
