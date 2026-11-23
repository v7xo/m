//go:build generate

package echo

//go:generate protoc --proto_path=. --proto_path=../../../../.pjcli/local/include --go-grpc_out=. --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false echo.proto
