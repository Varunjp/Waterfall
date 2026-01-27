package executor

import (
	"context"
	"fmt"
	"runtime/debug"

	"go.uber.org/zap"
)

type Executor interface {
	Execute(ctx context.Context, payload []byte)error 
}

var Registry = map[string]Executor{}

func Register(jobType string,exec Executor) {
	Registry[jobType] = exec
}

func SafeExecute(
	ctx context.Context,
	exec func(context.Context)error,
	log *zap.Logger,
)(err error) {
	defer func(){
		if r := recover(); r != nil {
			log.Error("job panicked",
				zap.Any("panic",r),
				zap.ByteString("stack",debug.Stack()),
			)
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return exec(ctx)
}