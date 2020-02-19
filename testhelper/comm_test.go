package testhelper

import (
	"loadgen/lib"
	"testing"
	"time"
)

func TestTCPComm(t *testing.T) {
	addr := "127.0.0.1:8000"
	stopSignal := make(chan struct{})
	//客户端
	go func(signal chan<- struct{}, addr string) {
		//等待1s让服务端先运行
		time.Sleep(1 * time.Second)

		comm := NewTCPComm(addr)
		rawReq := comm.BuildReq()

		start := time.Now().UnixNano()
		mBytes, err := comm.Call(rawReq.Req, 50*time.Millisecond)
		end := time.Now().UnixNano()

		result := comm.CheckResp(rawReq, lib.RawResp{
			ID:     rawReq.ID,
			Resp:   mBytes,
			Err:    err,
			Elapse: time.Duration(end - start),
		})
		lib.StdLog.Infof("result: %v\n", result)
		signal <- struct{}{}
	}(stopSignal, addr)
	//服务端
	server := NewTCPServer()
	if err := server.Listen(addr); err != nil {
		t.Fatal(err)
	}
	<-stopSignal
	server.Close()
}
