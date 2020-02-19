package loadgen

import (
	"testing"
	"time"

	loadgenlib "loadgen/lib"
	helper "loadgen/testhelper"
)

// 代表是否打印详细结果
var printDetail = false

func TestStart(t *testing.T) {
	// 初始化服务器
	server := helper.NewTCPServer()
	defer server.Close()
	serverAddr := "127.0.0.1:8000"
	t.Logf("Startup TCP Server(%s)...\n", serverAddr)
	err := server.Listen(serverAddr)
	if err != nil {
		t.Fatalf("TCP Server startup failing!(addr=%s)!\n", serverAddr)
	}

	// 初始化载荷器
	ps := ParamSet{
		Caller:     helper.NewTCPComm(serverAddr),
		TimeoutNS:  50 * time.Microsecond,
		LPS:        uint32(1000),
		DurationNS: 10 * time.Second,
		ResultCh:   make(chan *loadgenlib.CallResult, 50),
	}

	t.Logf("Initialize load generator (timeoutNS=%v, lps=%d, durationNS=%v)...",
		ps.TimeoutNS, ps.LPS, ps.DurationNS)
	gen, err := NewGenerator(ps)
	if err != nil {
		t.Fatalf("Load generator initalization failing:%s\n", err)
	}

	//开始
	t.Log("Start load generator")
	gen.Start()

	//显示结果
	countMap := make(map[loadgenlib.RetCode]int)
	for r := range ps.ResultCh {
		countMap[r.Code] = countMap[r.Code] + 1
		if printDetail {
			t.Logf("Result: %v", r)
		}
	}

	var total int
	t.Log("RetCode Count:")
	for k, v := range countMap {
		codePlain := loadgenlib.GetRetCodePlain(k)
		t.Logf("	Code plain:%s (%d), Count:%d.\n", codePlain, k, v)
		total += v
	}
	t.Logf("Total: %d.\n", total)
	successCount := countMap[loadgenlib.RET_CODE_SUCCESS]
	tps := float64(successCount) / float64(ps.DurationNS/1e9)
	t.Logf("Loads per second: %d; Treatments per second: %f.\n", ps.LPS, tps)
}

func TestStop(t *testing.T) {
	//初始化服务器
	server := helper.NewTCPServer()
	defer server.Close()
	serverAddr := "127.0.0.1:8081"
	t.Logf("Startup TCP server(%s)...\n", serverAddr)
	err := server.Listen(serverAddr)
	if err != nil {
		t.Fatalf("TCP Server startup failing!(addr=%s)!\n", serverAddr)
	}
	//初始化载荷发生器
	ps := ParamSet{
		Caller:     helper.NewTCPComm(serverAddr),
		TimeoutNS:  50 * time.Millisecond,
		LPS:        uint32(1000),
		DurationNS: 10 * time.Second,
		ResultCh:   make(chan *loadgenlib.CallResult, 50),
	}
	t.Logf("Initialize load generator (timeoutNS=%v, lps=%d, durationNS=%v)...",
		ps.TimeoutNS, ps.LPS, ps.DurationNS)
	gen, err := NewGenerator(ps)
	if err != nil {
		t.Fatalf("Load generator initialization failing:%s.\n", err)
	}
	//开始
	t.Log("Start load generator...")
	gen.Start()
	timeoutNS := 2 * time.Second
	time.AfterFunc(timeoutNS, func() {
		// Stop调用会关闭ResultCh
		gen.Stop()
	})
	//显示调用结果
	countMap := make(map[loadgenlib.RetCode]int)
	count := 0
	for r := range ps.ResultCh {
		countMap[r.Code] = countMap[r.Code] + 1
		if printDetail {
			t.Logf("result:%v\n", r)
		}
		count++
	}
	var total int
	t.Log("RetCode Count:")
	for k, v := range countMap {
		codePlain := loadgenlib.GetRetCodePlain(k)
		t.Logf("	Code Plain: %s (%d), Count: %d.\n", codePlain, k, v)
		total += v
	}
	t.Logf("Total: %d.\n", total)
	successCount := countMap[loadgenlib.RET_CODE_SUCCESS]
	tps := float64(successCount) / float64(timeoutNS/1e9)
	t.Logf("Loads per second: %d; Treatments per second: %f.\n", ps.LPS, tps)
}
