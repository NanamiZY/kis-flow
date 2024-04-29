package log

import (
	"context"
	"fmt"
)

type kisDefaultLog struct{}

func (log *kisDefaultLog) InfoF(str string, v ...interface{}) {
	fmt.Printf(str, v...)
	fmt.Printf("\n")
}

func (log *kisDefaultLog) ErrorF(str string, v ...interface{}) {
	fmt.Printf(str, v...)
	fmt.Printf("\n")
}

func (log *kisDefaultLog) DebugF(str string, v ...interface{}) {
	fmt.Printf(str, v...)
	fmt.Printf("\n")
}

func (log *kisDefaultLog) InfoFX(ctx context.Context, str string, v ...interface{}) {
	fmt.Println(ctx)
	fmt.Printf(str, v...)
	fmt.Printf("\n")
}

func (log *kisDefaultLog) ErrorFX(ctx context.Context, str string, v ...interface{}) {
	fmt.Println(ctx)
	fmt.Printf(str, v...)
	fmt.Printf("\n")
}

func (log *kisDefaultLog) DebugFX(ctx context.Context, str string, v ...interface{}) {
	fmt.Println(ctx)
	fmt.Printf(str, v...)
	fmt.Printf("\n")
}

func init() {
	if Logger() == nil {
		SetLogger(&kisDefaultLog{})
	}
}
