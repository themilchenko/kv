build_node:
	go build -o ./bin/node ./cmd/storage/main.go

build_clusterctl:
	go build -o ./bin/clusterctl ./cmd/clusterctl/main.go

build: build_node build_clusterctl
