package message

//go:generate protoc -I .  --proto_path=proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:proto  proto/message.proto
