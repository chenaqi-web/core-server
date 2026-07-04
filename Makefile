test:
	protoc --proto_path=./docs/proto \
	--go_out=./internal/rpc \
	--go_opt=paths=source_relative \
	--go-grpc_out=./internal/rpc \
	--go-grpc_opt=paths=source_relative \
	./docs/proto/health.proto
