package loadgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"loadgen/lib"
)

//日志记录器
var logger = lib.DLogger()

// myGenerator 代表载荷发生器的实现类型
type myGenerator struct {
	caller      lib.Caller           // 调用器
	timeoutNS   time.Duration        // 处理超时时间,单位:纳秒
	lps         uint32               // 每秒载荷量
	durationNS  time.Duration        // 负载持续时间,单位:纳秒
	concurrency uint32               // 载荷并发量
	tickets     lib.GoTickets        // Goroutine票池
	ctx         context.Context      // 上下文
	cancelFunc  context.CancelFunc   // 取消函数
	callCount   int64                // 调用计数
	status      uint32               // 状态
	resultCh    chan *lib.CallResult // 调用结果通道
}

// NewGenerator 新建一个载荷发生器
func NewGenerator(ps ParamSet) (lib.Generator, error) {
	logger.Infoln("New a load generator...")
	if err := ps.Check(); err != nil {
		return nil, err
	}
	gen := &myGenerator{
		caller:     ps.Caller,
		timeoutNS:  ps.TimeoutNS,
		lps:        ps.LPS,
		durationNS: ps.DurationNS,
		status:     lib.STATUS_ORIGINAL,
		resultCh:   ps.ResultCh,
	}
	if err := gen.init(); err != nil {
		return nil, err
	}
	return gen, nil
}

// init 初始化载荷发生器
func (gen *myGenerator) init() error {
	var buf bytes.Buffer
	buf.WriteString("Initializing the load generator...")
	// 载荷的并发量 ≈ 载荷的响应超时时间 / 载荷的发送间隔
	var total64 = int64(gen.timeoutNS)/int64(1e9/gen.lps) + 1
	if total64 > math.MaxInt32 {
		total64 = math.MaxInt32
	}
	gen.concurrency = uint32(total64)
	tickets, err := lib.NewGoTickets(gen.concurrency)
	if err != nil {
		return err
	}
	gen.tickets = tickets

	buf.WriteString(fmt.Sprintf("Done. (concurrency=%d)", gen.concurrency))
	logger.Infoln(buf.String())
	return nil
}

// callOne 向载荷承受方发起一次调用
func (gen *myGenerator) callOne(rawReq *lib.RawReq) *lib.RawResp {
	atomic.AddInt64(&gen.callCount, 1)
	if rawReq == nil {
		return &lib.RawResp{ID: -1, Err: errors.New("Invalid raw request.")}
	}
	//计算调用时长
	start := time.Now().UnixNano()
	resp, err := gen.caller.Call(rawReq.Req, gen.timeoutNS)
	end := time.Now().UnixNano()
	elapseTime := time.Duration(end - start)

	var rawResp *lib.RawResp
	if err != nil {
		errMsg := fmt.Sprintf("Sync Call Error: %s.", err)
		rawResp = &lib.RawResp{
			ID:     rawReq.ID,
			Err:    errors.New(errMsg),
			Elapse: elapseTime,
		}
	} else {
		rawResp = &lib.RawResp{
			ID:     rawReq.ID,
			Resp:   resp,
			Elapse: elapseTime,
		}
	}
	return rawResp
}

// asyncCall 异步地调用承受方接口
func (gen *myGenerator) asyncCall() {
	gen.tickets.Take()
	//异步发起调用
	go func() {
		defer func() {
			//防止接口调用goroutine恐慌导致载荷器整体退出
			if p := recover(); p != nil {
				err, ok := interface{}(p).(error)
				var errMsg string
				if ok {
					errMsg = fmt.Sprintf("Async Call Panic! (error: %s)", err)
				} else {
					errMsg = fmt.Sprintf("Async Call Panic! (error: %s)", p)
				}
				logger.Errorln(errMsg)
				//发生恐慌设置致命错误结果
				result := &lib.CallResult{
					ID:   -1,
					Code: lib.RET_CODE_FATAL_CALL,
					Msg:  errMsg,
				}
				gen.sendResult(result)
			}
			//归还票池
			gen.tickets.Return()
		}()
		//构建请求
		rawReq := gen.caller.BuildReq()
		var callStatus uint32
		//设定超时以及后续处理
		timer := time.AfterFunc(gen.timeoutNS, func() {
			//func是一个Goroutine处理, 这里要用一个原子操作对callStatus处理
			if !atomic.CompareAndSwapUint32(&callStatus, 0, 2) {
				return
			}
			//如果超时,将code设为TIMEOUT,接口调用耗时设为timeoutNS(载荷器限定的超时时间)
			result := &lib.CallResult{
				ID:     rawReq.ID,
				Req:    rawReq,
				Code:   lib.RET_CODE_WARNING_CALL_TIMEOUT,
				Msg:    fmt.Sprintf("Timeout! (expected: < %v)", gen.timeoutNS),
				Elapse: gen.timeoutNS,
			}
			//发送处理结果
			gen.sendResult(result)
		})
		//发送调用请求
		rawResp := gen.callOne(&rawReq)
		if !atomic.CompareAndSwapUint32(&callStatus, 0, 1) {
			return
		}
		timer.Stop()
		//正常来说,指不发生内部调用出错,resp的Elapse和result的Elapse是一致的
		var result *lib.CallResult
		if rawResp.Err != nil {
			result = &lib.CallResult{
				ID:     rawResp.ID,
				Req:    rawReq,
				Code:   lib.RET_CODE_ERROR_CALL,
				Msg:    rawResp.Err.Error(),
				Elapse: rawResp.Elapse,
			}
		} else {
			result = gen.caller.CheckResp(rawReq, *rawResp)
			result.Elapse = rawResp.Elapse
		}
		gen.sendResult(result)
	}()
}

// sendResult 用于发送处理调结果
func (gen *myGenerator) sendResult(result *lib.CallResult) bool {
	if atomic.LoadUint32(&gen.status) != lib.STATUS_STARTED {
		//未启动,打印结果忽略
		gen.printIgnoredResult(result, "stopped load generator")
		return false
	}
	select {
	case gen.resultCh <- result:
		return true
	default:
		//默认也对打印结果忽略
		gen.printIgnoredResult(result, "full result channel")
		return false
	}
}

// printIgnoredResult 打印忽略的结果
func (gen *myGenerator) printIgnoredResult(result *lib.CallResult, cause string) {
	resultMsg := fmt.Sprintf("ID=%d, Code=%d, Msg=%s, Elapse=%v", result.ID, result.Code, result.Msg, result.Elapse)
	logger.Warnf("Ignored result: %s. (cause: %s)", resultMsg, cause)
}

// prepareToStop 用于停止载荷发生做准备
func (gen *myGenerator) prepareToStop(ctxError error) {
	logger.Infof("Prepare to stop load generator (cuase: %s)...", ctxError)
	atomic.CompareAndSwapUint32(&gen.status, lib.STATUS_STARTED, lib.STATUS_STOPPING)
	close(gen.resultCh)
	atomic.StoreUint32(&gen.status, lib.STATUS_STOPPED)
}

// genLoad 产生载荷并向承受方发送
func (gen *myGenerator) genLoad(throttle <-chan time.Time) {
	for {
		select {
		case <-gen.ctx.Done():
			gen.prepareToStop(gen.ctx.Err())
			return
		default:
		}
		//异步发起载荷请求
		gen.asyncCall()
		if gen.lps > 0 {
			// select语句是伪随机,当节流阀的到期通知和上下文的信号同时到达,for语句开头
			// 再进行一次上下文, 确保载荷器及时退出
			select {
			case <-throttle:
			case <-gen.ctx.Done():
				gen.prepareToStop(gen.ctx.Err())
				return
			}
		}
	}
}

// Start 启动载荷发生器
func (gen *myGenerator) Start() bool {
	logger.Infoln("Starting load generator...")
	//检查是否具备可启的状态,顺便设置状态为正在启动
	if !atomic.CompareAndSwapUint32(&gen.status, lib.STATUS_ORIGINAL, lib.STATUS_STARTING) {
		if !atomic.CompareAndSwapUint32(&gen.status, lib.STATUS_STOPPED, lib.STATUS_STARTING) {
			return false
		}
	}

	//设定节流阀
	var throttle <-chan time.Time
	if gen.lps > 0 {
		interval := time.Duration(1e9 / gen.lps)
		logger.Infof("Setting throttle (%v)...", interval)
		throttle = time.Tick(interval)
	}

	//初始化上下文和取消函数
	gen.ctx, gen.cancelFunc = context.WithTimeout(context.Background(), gen.durationNS)

	//初始化调用计数
	gen.callCount = 0

	//设置状态为已启动
	atomic.StoreUint32(&gen.status, lib.STATUS_STARTED)

	go func() {
		//生成并发送载荷
		logger.Infoln("Generating loads...")
		gen.genLoad(throttle)
		logger.Infof("Stopped. (call count: %d)", gen.callCount)
	}()
	return false
}

// Stop 停止载荷发生器
func (gen *myGenerator) Stop() bool {
	if !atomic.CompareAndSwapUint32(&gen.status, lib.STATUS_STARTED, lib.STATUS_STOPPING) {
		return false
	}
	gen.cancelFunc()
	for {
		if atomic.LoadUint32(&gen.status) == lib.STATUS_STOPPED {
			break
		}
		time.Sleep(time.Microsecond)
	}
	return true
}

// Status 获取载荷器当前状态
func (gen *myGenerator) Status() uint32 {
	return atomic.LoadUint32(&gen.status)
}

// CallCount 获取载荷器调用计数
func (gen *myGenerator) CallCount() int64 {
	return atomic.LoadInt64(&gen.callCount)
}
