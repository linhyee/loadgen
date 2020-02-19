package testhelper

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"

	loadgenlib "loadgen/lib"
)

const (
	DELIM = '\n' //分隔附
)

// 代表操作符切片
var operators = []string{"+", "-", "*", "/"}

// TCPComm 表示TCP通迅器的结构
type TCPComm struct {
	addr string // 通讯地址
}

// NewTCPComm 新建一个TCP通讯器
func NewTCPComm(addr string) loadgenlib.Caller {
	return &TCPComm{addr: addr}
}

// BuildReq 构建一个请求
func (comm *TCPComm) BuildReq() loadgenlib.RawReq {
	//设置伪随机种子
	rand.Seed(time.Now().UnixNano())
	id := time.Now().UnixNano()
	sreq := ServerReq{
		ID: id,
		Operands: []int{
			int(rand.Int31n(1000) + 1),
			int(rand.Int31n(1000) + 1),
		},
		Operator: func() string {
			return operators[rand.Int31n(100)%4]
		}(),
	}
	mBytes, err := json.Marshal(sreq)
	if err != nil {
		panic(err)
	}
	rawReq := loadgenlib.RawReq{ID: id, Req: mBytes}
	return rawReq
}

// Call 发起一次通讯
func (comm *TCPComm) Call(req []byte, timeoutNS time.Duration) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", comm.addr, timeoutNS)
	if err != nil {
		return nil, err
	}
	_, err = write(conn, req, DELIM)
	if err != nil {
		return nil, err
	}
	return read(conn, DELIM)
}

// CheckResp 检查响应内容
func (comm *TCPComm) CheckResp(rawReq loadgenlib.RawReq, rawResp loadgenlib.RawResp) *loadgenlib.CallResult {
	var commResult loadgenlib.CallResult
	commResult.ID = rawReq.ID
	commResult.Req = rawReq
	commResult.Resp = rawResp

	var sReq ServerReq
	if err := json.Unmarshal(rawReq.Req, &sReq); err != nil {
		commResult.Code = loadgenlib.RET_CODE_FATAL_CALL
		commResult.Msg = fmt.Sprintf("Incorrectly formatted Req: %s!\n", string(rawReq.Req))
		return &commResult
	}

	var sResp ServerResp
	if err := json.Unmarshal(rawResp.Resp, &sResp); err != nil {
		commResult.Code = loadgenlib.RET_CODE_ERROR_RESPONSE
		commResult.Msg = fmt.Sprintf("Incorrectly formated Resp: %s!\n", string(rawResp.Resp))
		return &commResult
	}

	if sResp.ID != sReq.ID {
		commResult.Code = loadgenlib.RET_CODE_ERROR_RESPONSE
		commResult.Msg = fmt.Sprintf("Incorrectly raw id! (%d!=%d)\n", rawReq.ID, rawResp.ID)
		return &commResult
	}

	if sResp.Err != nil {
		commResult.Code = loadgenlib.RET_CODE_ERROR_CALEE
		commResult.Msg = fmt.Sprintf("Abnormal server: %s!\n", sResp.Err)
		return &commResult
	}

	if sResp.Result != op(sReq.Operands, sReq.Operator) {
		commResult.Code = loadgenlib.RET_CODE_ERROR_RESPONSE
		commResult.Msg = fmt.Sprintf("Incorrect result: %s!\n",
			genFormula(sReq.Operands, sReq.Operator, sResp.Result, false))
		return &commResult
	}

	commResult.Code = loadgenlib.RET_CODE_SUCCESS
	commResult.Msg = fmt.Sprintf("Success. (%s)", sResp.Formula)

	return &commResult
}

// read 从连接中读数据直到遇到参数delim代表的字节
func read(conn net.Conn, delim byte) ([]byte, error) {
	readBytes := make([]byte, 1)
	var buffer bytes.Buffer
	for {
		_, err := conn.Read(readBytes)
		if err != nil {
			return nil, err
		}
		readByte := readBytes[0]
		if readByte == delim {
			break
		}
		buffer.WriteByte(readByte)
	}
	return buffer.Bytes(), nil
}

// write 向连接写数据,并在最后追加参数delim代表的字节
func write(conn net.Conn, content []byte, delim byte) (int, error) {
	writer := bufio.NewWriter(conn)
	n, err := writer.Write(content)
	if err == nil {
		_ = writer.WriteByte(delim)
	}
	if err == nil {
		err = writer.Flush()
	}
	return n, err
}
