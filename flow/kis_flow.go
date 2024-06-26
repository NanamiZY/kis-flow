package flow

import (
	"KisFlow/common"
	"KisFlow/config"
	"KisFlow/function"
	"KisFlow/id"
	"KisFlow/kis"
	"KisFlow/log"
	"context"
	"errors"
	"sync"
)

// kisFlow 用于贯穿整条流式计算的上下文环境
type KisFlow struct {
	//基础信息
	Id   string                // Flow的分布式实例ID(用于KisFlow内部区分不同实例)
	Name string                // Flow的可读名称
	Conf *config.KisFlowConfig //Flow配置策略

	//Function列表
	Funcs          map[string]kis.Function //当前flow拥有的全部管理的全部Function对象,key:FunctionID
	FlowHead       kis.Function            //当前Flow所拥有的Function列表表头
	FlowTail       kis.Function            //当前Flow所拥有的Function列表表尾
	flock          sync.RWMutex            //管理链表插入读写的锁
	ThisFunction   kis.Function            //Flow当前正在执行的KisFunction对象
	ThisFunctionId string                  // 当前执行到的Function ID(策略配置ID)
	PrevFunctionId string                  //当前执行到的Function上一层FunctionID(策略配置ID)
	//Function列表参数
	funcParams map[string]config.FParam //flow在当前Function的自定义固定配置参数,Key:function的实例NsID, value:FParam
	fplock     sync.RWMutex             //管理funcParams的读写锁

	buffer common.KisRowArr  //用来临时存放输入字节数据的内部Buf,一条数据为interface{},多条数据为[]interface{},即KisBatch
	data   common.KisDataMap //流式计算各个层级的数据源
	inPut  common.KisRowArr  //当前Function的计算输入数据
}

// NewKisFlow 创建一个KisFlow.
func NewKisFlow(conf *config.KisFlowConfig) kis.Flow {
	flow := new(KisFlow)
	//实例Id
	flow.Id = id.KisID(common.KisIdTypeFlow)

	// 基础信息
	flow.Name = conf.FlowName
	flow.Conf = conf

	//Function列表
	flow.Funcs = make(map[string]kis.Function)
	flow.funcParams = make(map[string]config.FParam)

	flow.data = make(common.KisDataMap)

	return flow
}

// Link 将Function链接到Flow中
// fConf: 当前Function策略
// fParams:当前Flow携带的Function动态参数
func (flow *KisFlow) Link(fConf *config.KisFuncConfig, fParams config.FParam) error {
	//创建Function
	f := function.NewKisFunction(flow, fConf)

	//Flow添加Function
	if err := flow.appendFunc(f, fParams); err != nil {
		return err
	}
	return nil
}

// appendFunc 将Function添加到Flow中，链表操作
func (flow *KisFlow) appendFunc(function kis.Function, fParam config.FParam) error {
	if function == nil {
		return errors.New("AppendFunc append nil to List")
	}
	flow.flock.Lock()
	defer flow.flock.Unlock()

	if flow.FlowHead == nil {
		//首次添加节点
		flow.FlowHead = function
		flow.FlowTail = function

		function.SetN(nil)
		function.SetP(nil)
	} else {
		//将function插入到链表的尾部
		function.SetP(flow.FlowTail)
		function.SetN(nil)

		flow.FlowTail.SetN(function)
		flow.FlowTail = function
	}

	//将Function ID 详细Hash对应关系添加到flow对象中
	flow.Funcs[function.GetId()] = function

	//先添加function默认携带的Params参数
	params := make(config.FParam)
	for key, value := range function.GetConfig().Option.Params {
		params[key] = value
	}

	//再添加flow携带的function定义参数(重复即覆盖)
	for key, value := range fParam {
		params[key] = value
	}
	// 将得到的FParams存留在flow结构体中，用来function业务直接通过Hash获取
	//key为当前Function的KisId,不用Fid的原因是为了防止一个Flow添加两个相同策略Id的Function
	flow.funcParams[function.GetId()] = params
	return nil
}

// Run启动KisFlow的流式计算，从起始Function开始执行流
func (flow *KisFlow) Run(ctx context.Context) error {
	var fn kis.Function
	fn = flow.FlowHead
	if flow.Conf.Status == int(common.FlowDisable) {
		//flow被配置关闭
		return nil
	}

	flow.PrevFunctionId = common.FunctionIdFirstVirtual

	//提交数据流原始数据
	if err := flow.commitSrcData(ctx); err != nil {
		return err
	}
	//流式链式调用
	for fn != nil {
		// flow记录当前执行到的Function标记
		fid := fn.GetId()
		if flow.ThisFunctionId != "" {
			flow.PrevFunctionId = flow.ThisFunctionId
		}
		flow.ThisFunction = fn
		flow.ThisFunctionId = fid

		//得到当前Function要处理的源数据
		if inputData, err := flow.getCurData(); err != nil {
			log.Logger().ErrorFX(ctx, "flow.Run(): getCurData err = %s\n", err.Error())
			return err
		} else {
			flow.inPut = inputData
		}
		if err := fn.Call(ctx, flow); err != nil {
			//Error
			return err
		} else {
			//Success
			if err := flow.commitCurData(ctx); err != nil {
				return err
			}
			fn = fn.Next()
		}
	}
	return nil
}

func (flow *KisFlow) GetName() string {
	return flow.Name
}

func (flow *KisFlow) GetThisFunction() kis.Function {
	return flow.ThisFunction
}

func (flow *KisFlow) GetThisFuncConf() *config.KisFuncConfig {
	return flow.ThisFunction.GetConfig()
}
