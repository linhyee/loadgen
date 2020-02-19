package testhelper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"loadgen/lib"
)

// 日志记录器
var logger = lib.DLogger()

// ServerReq 表示服务器请求的结构
type ServerReq struct {
	ID       int64
	Operands []int
	Operator string
}

// ServerResp 表示服务器响应的结构
type ServerResp struct {
	ID      int64
	Formula string
	Result  int
	Err     error
}

// op 运算操作
func op(operands []int, operator string) int {
	var result int
	switch operator {
	case "+":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result += v
			}
		}
	case "-":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result -= v
			}
		}
	case "*":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result *= v
			}
		}
	case "/":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result /= v
			}
		}
	}
	return result
}

// genFormula 根据参数生成字符串形式的公式
func genFormula(operands []int, operator string, result int, equal bool) string {
	var buff bytes.Buffer
	n := len(operands)
	for i := 0; i < n; i++ {
		if i > 0 {
			buff.WriteString(" ")
			buff.WriteString(operator)
			buff.WriteString(" ")
		}
		buff.WriteString(strconv.Itoa(operands[i]))
	}
	if equal {
		buff.WriteString(" = ")
	} else {
		buff.WriteString(" != ")
	}
	buff.WriteString(strconv.Itoa(result))
	return buff.String()
}

// reqHandler 把参数sResp代表的请求转换为数据并发送给连接的客户端
func reqHandler(conn net.Conn) {
	var errMsg string
	var sResp ServerResp
	req, err := read(conn, DELIM)
	if err != nil {
		errMsg = fmt.Sprintf("Server: Req Read Error:%s", err)
	} else {
		var sReq ServerReq
		err := json.Unmarshal(req, &sReq)
		if err != nil {
			errMsg = fmt.Sprintf("Server: Req Unmarshal Error:%s", err)
		} else {
			sResp.ID = sReq.ID
			sResp.Result = op(sReq.Operands, sReq.Operator)
			sResp.Formula = genFormula(sReq.Operands, sReq.Operator, sResp.Result, true)
		}
	}
	if errMsg != "" {
		sResp.Err = errors.New(errMsg)
	}
	mBytes, err := json.Marshal(sResp)
	if err != nil {
		logger.Errorf("Server: Resp Marshal Error:%s", err)
	}
	_, err = write(conn, mBytes, DELIM)
	if err != nil {
		logger.Errorf("Server: Resp Write Error: %s", err)
	}
}

// TCPServer 表示基于TCP协议的服务器
type TCPServer struct {
	listener net.Listener
	active   uint32 // 0-未激活;2-已激活
}

// NewTCPServer 新建一个基于TCP协议的服务器
func NewTCPServer() *TCPServer {
	return &TCPServer{}
}

// init 初始化服务器
func (s *TCPServer) init(addr string) error {
	if !atomic.CompareAndSwapUint32(&s.active, 0, 1) {
		return nil
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		atomic.StoreUint32(&s.active, 0)
		return err
	}
	s.listener = ln
	return nil
}

// Listen 启动对指定网络地址的监听
func (s *TCPServer) Listen(addr string) error {
	err := s.init(addr)
	if err != nil {
		return err
	}
	go func() {
		for {
			if atomic.LoadUint32(&s.active) != 1 {
				break
			}
			conn, err := s.listener.Accept()
			if err != nil {
				if atomic.LoadUint32(&s.active) == 1 {
					logger.Errorf("Server: Request Acceptation Error: %s\n", err)
				} else {
					logger.Warnf("server: Broken Acceptation because of closed network connection:%s.\n", err)
				}
				continue
			}
			go reqHandler(conn)
		}
	}()
	return nil
}

// Close 关闭服务器
func (s *TCPServer) Close() bool {
	if !atomic.CompareAndSwapUint32(&s.active, 1, 0) {
		return false
	}
	_ = s.listener.Close()
	return true
}
