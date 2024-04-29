package test

import (
	"KisFlow/common"
	"KisFlow/config"
	"KisFlow/flow"
	"KisFlow/function"
	"context"
	"testing"
)

func TestNewKisFunction(t *testing.T) {
	ctx := context.Background()

	// 1. 创建一个KisFunction配置实例
	source := config.KisSource{
		Name: "商城用户订单数据",
		Must: []string{"order_id", "user_id"},
	}
	myFuncConfig1 := config.NewFuncConfig("funcName1", common.C, &source, nil)
	if myFuncConfig1 == nil {
		panic("myFuncConfig is nil")
	}

	// 2. 创建一个KisFlow配置实例
	myFlowConfig1 := config.NewFlowConfig("flowName1", common.FlowEnable)

	// 3. 创建一个KisFlow对象
	flow1 := flow.NewKisFlow(myFlowConfig1)

	// 4. 创建一个KisFunction对象
	func1 := function.NewKisFunction(flow1, myFuncConfig1)

	if err := func1.Call(ctx, flow1); err != nil {
		t.Errorf("func1.Call() error = %v", err)
	}
}
