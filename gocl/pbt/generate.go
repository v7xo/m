//go:build generate

package protobuf

//go:generate protoc --proto_path=protobuf --proto_path ../../.pjcli/local/include --go_out=. --go_opt=module=rn.m.vo7.io/vo7io/m/gocl/pbt protobuf/uuid.proto protobuf/timestamp.proto protobuf/pagination.proto protobuf/error.proto protobuf/echo.proto protobuf/filter.proto
