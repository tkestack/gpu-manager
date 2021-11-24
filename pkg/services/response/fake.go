package response

import (
	"k8s.io/klog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type fakeResponseManager struct {
	data map[string]containerResponseDataMapping
}

var _ Manager = (*fakeResponseManager)(nil)

func NewFakeResponseManager() *fakeResponseManager {
	return &fakeResponseManager{
		data: make(map[string]containerResponseDataMapping),
	}
}

func (m *fakeResponseManager) LoadFromFile(path string) error {
	return nil
}

func (m *fakeResponseManager) InsertResp(podUID, containerName string, allocResp *pluginapi.ContainerAllocateResponse) {
	podData, ok := m.data[podUID]
	if !ok {
		podData = make(containerResponseDataMapping)
		m.data[podUID] = podData
	}

	podData[containerName] = allocResp

	klog.V(2).Infof("Insert %s/%s allocResp", podUID, containerName)
}

func (m *fakeResponseManager) DeleteResp(podUID string, containerName string) {
	podData, ok := m.data[podUID]
	if !ok {
		return
	}

	_, ok = podData[containerName]
	if !ok {
		return
	}

	klog.V(2).Infof("Delete %s/%s allocResp", podUID, containerName)

	delete(podData, containerName)

	if len(podData) == 0 {
		delete(m.data, podUID)
	}
}

func (m *fakeResponseManager) GetResp(podUID string, containerName string) *pluginapi.ContainerAllocateResponse {
	podData, ok := m.data[podUID]
	if !ok {
		return nil
	}

	resp, ok := podData[containerName]
	if !ok {
		return nil
	}

	return resp
}

func (m *fakeResponseManager) ListAll() map[string]containerResponseDataMapping {
	return m.data
}
