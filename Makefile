.PHONY: generate-proto-auth

generate-proto-auth: 
	protoc -I protos/proto protos/proto/auth/auth.proto --go_out=protos/gen/go --go_opt=paths=source_relative --go-grpc_out=protos/gen/go --go-grpc_opt=paths=source_relative

.PHONY: generate-proto-queue
generate-proto-queue: 
	protoc -I protos/proto protos/proto/queue/queue.proto --go_out=protos/gen/go --go_opt=paths=source_relative --go-grpc_out=protos/gen/go --go-grpc_opt=paths=source_relative

.PHONY: generate-proto-notification
generate-proto-notification: 
	protoc -I protos/proto protos/proto/notification/notification.proto --go_out=protos/gen/go --go_opt=paths=source_relative --go-grpc_out=protos/gen/go --go-grpc_opt=paths=source_relative
