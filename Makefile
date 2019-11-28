.PHONY: all
all:
	hack/build.sh manager client

.PHONY: clean
clean:
	rm -rf ./go

.PHONY: test
test:
	hack/build.sh "test"

.PHONY: proto
proto:
	hack/build.sh "proto"

.PHONY: img
img:
	hack/build.sh "img"

.PHONY: fmt
fmt:
	hack/build.sh "fmt"

.PHONY: lint
lint:
	@revive -config revive.toml -exclude vendor/... -exclude pkg/api/runtime/... ./...
