package runtime

type ContainerInfo struct {
	Name         string
	ID           string
	Labels       map[string]string
	CgroupParent string
}

type ContainerRuntimeInterface interface {
	// CgroupDriver returns the cgroup driver of container runtime
	CgroupDriver() (string, error)
	// InspectContainer returns the container information by the given name
	InspectContainer(name string) (*ContainerInfo, error)
}

const (
	DockerRuntime = "docker"
	CRIORuntime   = "cri-o"
)
