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

type WriterSyncer interface {
	io.Writer
	Sync() error
}

func Init(wrt WriterSyncer, debug bool) {
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

func Close() {
	baseLogger.Sync()
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}
