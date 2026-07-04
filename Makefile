MODULE := backend/core-server
PROTO_DIR := ./docs/proto
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

wire:
	go run github.com/google/wire/cmd/wire ./cmd/server

generate-proto-rpc:
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=. \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(MODULE) \
		$(PROTO_FILES)
