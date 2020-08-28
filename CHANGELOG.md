<a name="unreleased"></a>
## [Unreleased]

### Feat
- Upgrade vcuda-controller to v1.0.2

### Fix
- wait server until it's ready


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

[Unreleased]: https://github.com/tkestack/gpu-manager/compare/v1.0.4...HEAD
[v1.0.4]: https://github.com/tkestack/gpu-manager/compare/v1.1.0...v1.0.4
[v1.1.0]: https://github.com/tkestack/gpu-manager/compare/v1.0.3...v1.1.0
