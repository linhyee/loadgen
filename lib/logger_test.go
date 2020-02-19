package lib

import (
	"testing"
)

func TestStdLog(t *testing.T) {

	//测试 默认debug输出
	Debug("log debug content1")
	Debug("log debug content2")

	Debugf(" log debug a = %d\n", 10)

	//设置log标记位，加上长文件名称 和 微秒 标记
	ResetFlags(BitDate | BitLongFile | BitLevel)
	Info("log info content")

	//设置日志前缀，主要标记当前日志模块
	SetPrefix("MODULE")
	Error("log error content")

	//添加标记位
	AddFlag(BitShortFile | BitTime)
	Stack(" log Stack! ")

	//设置日志写入文件
	SetLogFile("./log", "testfile.log")
	Debug("===> log debug content ~~666")
	Debug("===> log debug content ~~888")
	Error("===> log Error!!!! ~~~555~~~")

	//关闭debug调试
	CloseDebug()
	Debug("===> 我不应该出现~！")
	Debug("===> 我不应该出现~！")
	Error("===> log Error  after debug close !!!!")

}
