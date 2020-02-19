package lib

import "os"

// MyLogger 代表日志记录器的接口(原书接口)
type MyLogger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Errorln(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Infoln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Warnln(v ...interface{})
}

// AdapterLogger 适应原书日志组件的适配实例
type AdapterLogger struct {
	logger *Logger
}

func DLogger() MyLogger {
	return &AdapterLogger{logger: NewLog(os.Stderr, "", BitDefault)}
}

func (log *AdapterLogger) Debug(v ...interface{}) {
	log.logger.Debug(v...)
}

func (log *AdapterLogger) Debugf(format string, v ...interface{}) {
	log.logger.Debugf(format, v...)
}

func (log *AdapterLogger) Debugln(v ...interface{}) {
	log.logger.Debug(v...)
}

func (log *AdapterLogger) Error(v ...interface{}) {
	log.logger.Error(v...)
}

func (log *AdapterLogger) Errorf(format string, v ...interface{}) {
	log.logger.Errorf(format, v...)
}

func (log *AdapterLogger) Errorln(v ...interface{}) {
	log.logger.Error(v...)
}

func (log *AdapterLogger) Fatal(v ...interface{}) {
	log.logger.Fatal(v...)
}

func (log *AdapterLogger) Fatalf(format string, v ...interface{}) {
	log.logger.Fatalf(format, v...)
}

func (log *AdapterLogger) Fatalln(v ...interface{}) {
	log.logger.Fatal(v...)
}

func (log *AdapterLogger) Info(v ...interface{}) {
	log.logger.Info(v...)
}

func (log *AdapterLogger) Infof(format string, v ...interface{}) {
	log.logger.Infof(format, v...)
}

func (log *AdapterLogger) Infoln(v ...interface{}) {
	log.logger.Info(v...)
}

func (log *AdapterLogger) Panic(v ...interface{}) {
	log.logger.Panic(v...)
}

func (log *AdapterLogger) Panicf(format string, v ...interface{}) {
	log.logger.Panicf(format, v...)
}

func (log *AdapterLogger) Panicln(v ...interface{}) {
	log.logger.Panic(v...)
}

func (log *AdapterLogger) Warn(v ...interface{}) {
	log.logger.Warn(v...)
}

func (log *AdapterLogger) Warnf(format string, v ...interface{}) {
	log.logger.Warnf(format, v...)
}

func (log *AdapterLogger) Warnln(v ...interface{}) {
	log.logger.Warn(v...)
}
