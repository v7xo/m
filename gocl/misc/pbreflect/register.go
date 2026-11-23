package pbreflect

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

func RegisterProtoFilesGlobally(f []protoreflect.FileDescriptor) (err error) {
	return RegisterProtoFiles(protoregistry.GlobalFiles, f)
}

func RegisterProtoFiles(reg *protoregistry.Files, f []protoreflect.FileDescriptor) (err error) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			var (
				panicErr   error
				panicErrOk bool
			)

			if panicErr, panicErrOk = panicVal.(error); !panicErrOk {
				panicErr = fmt.Errorf("panic error value: %v", panicVal)
			}

			err = errs.Wrap(panicErr, err)
		}

		if err != nil {
			err = fmt.Errorf("failed to register .proto files for reflection; %w", err)
		}
	}()

	for i := range f {
		if err = reg.RegisterFile(f[i]); err != nil {
			return
		}
	}

	return
}
