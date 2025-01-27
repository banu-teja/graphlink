.PHONY: proto build run

proto:
	mkdir -p pkg/api/graph
	protoc --proto_path=internal/pb \
	    --go_out=pkg/api/graph --go_opt=paths=source_relative \
	    --go-grpc_out=pkg/api/graph --go-grpc_opt=paths=source_relative \
	    internal/pb/graph.proto

build: proto
	go build -o bin/graphlink cmd/server/main.go

run: build
	./bin/graphlink

.PHONY: docker-build docker-run
docker-build: build
	docker build -t graphlink .

docker-run: docker-build
	docker run -d -p 50051:50051 graphlink