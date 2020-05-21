# FAQ

*1.* Q: If I use another container runtime, what should I do?

A: You need to change the `EXTRA_FLAGS` of `gpu-manager.yaml`, add `--container-runtime-endpoint` options, the value is the
path of your container runtime unix socket, like `/var/run/crio.sock` or something like that.

*2.* Q: When I use a fraction gpu resource, my program hung

A: Add environment variable `LOGGER_LEVEL` and set value to `5` to `gpu-manager.yaml, and paste your log in your issue.

*3.* Q: When I use a fraction gpu resource, program reported a error like `rpc failed`

A: After v1.0.3, we use CRI interface to find cgroup path, so if your cgroup driver is not `cgroupfs`, you
need to change the `EXTRA_FLAGS` of `gpu-manager.yaml`, add `--cgroup-driver` options, the possible options are `cgroupfs` or `systemd`.
