<a name="unreleased"></a>
## [Unreleased]

### Feat
- add timeout option waiting for all resource server ready
- upgrade vcuda

### Fix
- virutal manager can't probe correct vm controller path
- wrong size of device memory when app has more than 1 card
- vcuda image repository url
- the mismatch between gpu-manager pick up and gpu-admission predicate ([#74](https://github.com/tkestack/gpu-manager/issues/74))
- read QoS class from pod status first
- kubelet 1.20 device checkpoint support ([#62](https://github.com/tkestack/gpu-manager/issues/62))
- report device memory when allocate more than one cards
- DeviceGetTopologyCommonAncestor get a zero value on multi-gpu board
- gpu-manager lost checkpoint data file
- preserve attribute
- read cgroup.procs files recursively
- wait server until it's ready


<a name="v1.1.5"></a>
## [v1.1.5] - 2021-05-10
### Docs
- Add FAQ link
- Update gpu manager yaml

### Feat
- upgrade vcuda to 1.0.3
- Upgrade vcuda-controller to v1.0.1
- Use host network to build image
- Update go version to 1.14.3
- Support CRI interface

### Fix
- kubelet 1.20 device checkpoint support ([#62](https://github.com/tkestack/gpu-manager/issues/62))
- the mismatch between gpu-manager pick up and gpu-admission predicate ([#74](https://github.com/tkestack/gpu-manager/issues/74))
- read QoS class from pod status first
- report device memory when allocate more than one cards
- gpu-manager lost checkpoint data file
- DeviceGetTopologyCommonAncestor get a zero value on multi-gpu board
- preserve attribute
- read cgroup.procs files recursively
- wait server until it's ready
- Revert using vendor directory
- Allow non-root user to communicate with gpu manager
- Change ius rpm broken link
- skip symlink when copy bin to |${NV_DIR}|. ([#15](https://github.com/tkestack/gpu-manager/issues/15))

### Refact
- Use vendor directory
- Refact gpu-manager code


<a name="v1.0.9"></a>
## [v1.0.9] - 2021-02-23
### Feat
- use apiserver cache to list pod

### Fix
- ignore not running container while recovering


<a name="v1.0.8"></a>
## [v1.0.8] - 2021-02-22
### Feat
- Upgrade vcuda-controller to v1.0.2
- Use host network to build image
- Upgrade vcuda-controller to v1.0.1

### Fix
- missing recover tree data if information is retrieved from checkpoint file
- gpu-manager lost checkpoint data file
- DeviceGetTopologyCommonAncestor get a zero value on multi-gpu board
- preserve attribute
- upgrade go to 1.15
- wait server until it's ready
- Change ius rpm broken link
- Allow non-root user to communicate with gpu manager

### Refact
- only watch pod belong this node


<a name="v1.1.4"></a>
## [v1.1.4] - 2021-02-05
### Fix
- read QoS class from pod status first


<a name="v1.1.3"></a>
## [v1.1.3] - 2021-02-02
### Feat
- upgrade vcuda to 1.0.3

### Fix
- report device memory when allocate more than one cards


<a name="v1.1.2"></a>
## [v1.1.2] - 2020-12-09
### Docs
- Add FAQ link
- Update gpu manager yaml

### Feat
- Upgrade vcuda-controller to v1.0.1
- Use host network to build image
- Update go version to 1.14.3
- Support CRI interface

### Fix
- gpu-manager lost checkpoint data file
- DeviceGetTopologyCommonAncestor get a zero value on multi-gpu board
- preserve attribute
- read cgroup.procs files recursively
- wait server until it's ready
- Revert using vendor directory
- Allow non-root user to communicate with gpu manager
- Change ius rpm broken link
- skip symlink when copy bin to |${NV_DIR}|. ([#15](https://github.com/tkestack/gpu-manager/issues/15))

### Refact
- Use vendor directory
- Refact gpu-manager code


<a name="v1.0.7"></a>
## [v1.0.7] - 2020-12-09
### Feat
- Upgrade vcuda-controller to v1.0.2
- Use host network to build image
- Upgrade vcuda-controller to v1.0.1

### Fix
- gpu-manager lost checkpoint data file
- DeviceGetTopologyCommonAncestor get a zero value on multi-gpu board
- preserve attribute
- upgrade go to 1.15
- wait server until it's ready
- Change ius rpm broken link
- Allow non-root user to communicate with gpu manager

### Refact
- only watch pod belong this node


<a name="v1.1.1"></a>
## [v1.1.1] - 2020-12-02
### Docs
- Add FAQ link
- Update gpu manager yaml

### Feat
- Upgrade vcuda-controller to v1.0.1
- Use host network to build image
- Update go version to 1.14.3
- Support CRI interface

### Fix
- DeviceGetTopologyCommonAncestor get a zero value on multi-gpu board
- preserve attribute
- read cgroup.procs files recursively
- wait server until it's ready
- Revert using vendor directory
- Allow non-root user to communicate with gpu manager
- Change ius rpm broken link
- skip symlink when copy bin to |${NV_DIR}|. ([#15](https://github.com/tkestack/gpu-manager/issues/15))

### Refact
- Use vendor directory
- Refact gpu-manager code


<a name="v1.0.6"></a>
## [v1.0.6] - 2020-12-02
### Fix
- DeviceGetTopologyCommonAncestor get a zero value on multi-gpu board
- preserve attribute


<a name="v1.0.5"></a>
## [v1.0.5] - 2020-08-28
### Feat
- Upgrade vcuda-controller to v1.0.2

### Fix
- upgrade go to 1.15
- wait server until it's ready

### Refact
- only watch pod belong this node


<a name="v1.0.4"></a>
## [v1.0.4] - 2020-05-21
### Feat
- Use host network to build image
- Upgrade vcuda-controller to v1.0.1

### Fix
- Change ius rpm broken link
- Allow non-root user to communicate with gpu manager


<a name="v1.1.0"></a>
## [v1.1.0] - 2020-05-21
### Docs
- Add FAQ link
- Update gpu manager yaml

### Feat
- Upgrade vcuda-controller to v1.0.1
- Use host network to build image
- Update go version to 1.14.3
- Support CRI interface

### Fix
- Revert using vendor directory
- Allow non-root user to communicate with gpu manager
- Change ius rpm broken link
- skip symlink when copy bin to |${NV_DIR}|. ([#15](https://github.com/tkestack/gpu-manager/issues/15))

### Refact
- Use vendor directory
- Refact gpu-manager code


<a name="v1.0.3"></a>
## v1.0.3 - 2019-12-17

[Unreleased]: https://github.com/tkestack/gpu-manager/compare/v1.1.5...HEAD
[v1.1.5]: https://github.com/tkestack/gpu-manager/compare/v1.0.9...v1.1.5
[v1.0.9]: https://github.com/tkestack/gpu-manager/compare/v1.0.8...v1.0.9
[v1.0.8]: https://github.com/tkestack/gpu-manager/compare/v1.1.4...v1.0.8
[v1.1.4]: https://github.com/tkestack/gpu-manager/compare/v1.1.3...v1.1.4
[v1.1.3]: https://github.com/tkestack/gpu-manager/compare/v1.1.2...v1.1.3
[v1.1.2]: https://github.com/tkestack/gpu-manager/compare/v1.0.7...v1.1.2
[v1.0.7]: https://github.com/tkestack/gpu-manager/compare/v1.1.1...v1.0.7
[v1.1.1]: https://github.com/tkestack/gpu-manager/compare/v1.0.6...v1.1.1
[v1.0.6]: https://github.com/tkestack/gpu-manager/compare/v1.0.5...v1.0.6
[v1.0.5]: https://github.com/tkestack/gpu-manager/compare/v1.0.4...v1.0.5
[v1.0.4]: https://github.com/tkestack/gpu-manager/compare/v1.1.0...v1.0.4
[v1.1.0]: https://github.com/tkestack/gpu-manager/compare/v1.0.3...v1.1.0
