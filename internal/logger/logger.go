package logger

import (
	"io"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var baseLogger *zap.Logger

var logger *zap.SugaredLogger

func init() {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, os.Stdout, zap.InfoLevel)

	baseLogger = zap.New(core, zap.WithCaller(false), zap.AddStacktrace(zap.ErrorLevel))
	logger = baseLogger.Sugar()
}

type writerSyncer interface {
	io.Writer
	Sync() error
}

//Init - initialize logger
// wrt - where need to write logs 
// debug - if true it set DebugLevel, if false it set InfoLevel 
func Init(wrt writerSyncer, debug bool) {
	Close()

	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(config)

	core := zapcore.NewCore(encoder, wrt, level)

	baseLogger = zap.New(core, zap.WithCaller(false), zap.AddStacktrace(zap.ErrorLevel))
	logger = baseLogger.Sugar()

	logger = logger.With("uuid", uuid.New().String())
}

//Close - flushes buffered logs
func Close() {
	baseLogger.Sync()
}

//Debug logs the provided arguments at [DebugLevel]
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

//Info logs the provided arguments at [InfoLevel]
func Info(args ...interface{}) {
	logger.Info(args...)
}

//Warn logs the provided arguments at [WarnLevel]
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

//Error logs the provided arguments at [ErrorLevel]
func Error(args ...interface{}) {
	logger.Error(args...)
}

//Debugf formats the message according to the format specifier and logs it at [DebugLevel]
func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

//Infof formats the message according to the format specifier and logs it at [InfoLevel]
func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

//Warnf formats the message according to the format specifier and logs it at [WarnLevel]
func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

//Errorf formats the message according to the format specifier and logs it at [ErrorLevel]
func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

//Debugw logs a message with some additional context. Context is key-value pairs 
func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

//Infow logs a message with some additional context. Context is key-value pairs 
func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

//Warnw logs a message with some additional context. Context is key-value pairs 
func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

//Errorw logs a message with some additional context. Context is key-value pairs 
func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}
