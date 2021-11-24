package response

import (
	"sync"

	"k8s.io/klog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"tkestack.io/gpu-manager/pkg/types"
	"tkestack.io/gpu-manager/pkg/utils"
)

type Manager interface {
	InsertResp(podUID, containerName string, resp *pluginapi.ContainerAllocateResponse)
	DeleteResp(podUID, containerName string)
	GetResp(podUID, containerName string) *pluginapi.ContainerAllocateResponse
	ListAll() map[string]containerResponseDataMapping
	LoadFromFile(path string) error
}

var _ Manager = (*responseManager)(nil)

type responseManager struct {
	l    sync.Mutex
	data map[string]containerResponseDataMapping
}

type containerResponseDataMapping map[string]*pluginapi.ContainerAllocateResponse

func NewResponseManager() *responseManager {
	return &responseManager{
		data: make(map[string]containerResponseDataMapping),
	}
}

func (m *responseManager) LoadFromFile(path string) error {
	cp, err := utils.GetCheckpointData(path)
	if err != nil {
		return err
	}

	for _, item := range cp.PodDeviceEntries {
		// Only vcore resource has valid response data
		if item.ResourceName == types.VCoreAnnotation {
			allocResp := &pluginapi.ContainerAllocateResponse{}
			if err := allocResp.Unmarshal(item.AllocResp); err != nil {
				return err
			}

			m.InsertResp(item.PodUID, item.ContainerName, allocResp)
		}
	}

	return nil
}

func (m *responseManager) InsertResp(podUID, containerName string, allocResp *pluginapi.ContainerAllocateResponse) {
	m.l.Lock()
	defer m.l.Unlock()

	podData, ok := m.data[podUID]
	if !ok {
		podData = make(containerResponseDataMapping)
		m.data[podUID] = podData
	}

	podData[containerName] = allocResp

	klog.V(2).Infof("Insert %s/%s allocResp", podUID, containerName)
}

func (m *responseManager) DeleteResp(podUID string, containerName string) {
	m.l.Lock()
	defer m.l.Unlock()

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

func (m *responseManager) GetResp(podUID string, containerName string) *pluginapi.ContainerAllocateResponse {
	m.l.Lock()
	defer m.l.Unlock()

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

func (m *responseManager) ListAll() map[string]containerResponseDataMapping {
	m.l.Lock()
	defer m.l.Unlock()

	snapshot := make(map[string]containerResponseDataMapping)
	for uid, containerMapping := range m.data {
		podData, ok := snapshot[uid]
		if !ok {
			podData = make(containerResponseDataMapping)
			snapshot[uid] = podData
		}

		for name, resp := range containerMapping {
			podData[name] = resp
		}
	}

	return snapshot
}
