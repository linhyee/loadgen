package loadgen

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"loadgen/lib"
)

// ParamSet 代表子载荷发生器参数的集合
type ParamSet struct {
	Caller     lib.Caller           // 调用器
	TimeoutNS  time.Duration        // 响应超时时间, 单位:纳秒
	LPS        uint32               // 每秒载荷数
	DurationNS time.Duration        // 负载持续时间, 单位:纳秒
	ResultCh   chan *lib.CallResult // 调用结果通道
}

// Check 检查当前值的所有字段的有效性
// 若存在无效字段则返回值非nil
func (ps *ParamSet) Check() error {
	var errMsgs []string
	if ps.Caller == nil {
		errMsgs = append(errMsgs, "Invalid caller!")
	}
	if ps.TimeoutNS == 0 {
		errMsgs = append(errMsgs, "Invalid timeoutNS!")
	}
	if ps.LPS == 0 {
		errMsgs = append(errMsgs, "Invalid lps(load per second)!")
	}
	if ps.DurationNS == 0 {
		errMsgs = append(errMsgs, "Invalid durationsNS!")
	}
	if ps.ResultCh == nil {
		errMsgs = append(errMsgs, "Invalid result channel!")
	}
	var buf bytes.Buffer
	buf.WriteString("Checking the parameters...")
	if errMsgs != nil {
		errMsg := strings.Join(errMsgs, " ")
		buf.WriteString(fmt.Sprintf("NOT passed! (%s)", errMsg))
		logger.Infoln(buf.String())
		return errors.New(errMsg)
	}
	buf.WriteString(
		fmt.Sprintf("Passed. (timeoutNS=%s, lps=%d, durationNS=%s)",
			ps.TimeoutNS, ps.LPS, ps.DurationNS))
	return nil
}
