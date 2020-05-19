.PHONY: all

all: 
	./build.sh

.PHONY: clean

clean:
	rm -rf ./go ./go-nvml

.PHONY: lint

lint:
	@revive -config revive.toml -exclude types.go ./...
