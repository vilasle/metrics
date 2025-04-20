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

//TODO add godoc
type WriterSyncer interface {
	io.Writer
	Sync() error
}

//TODO add godoc
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

//TODO add godoc
func Close() {
	baseLogger.Sync()
}

//TODO add godoc
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

//TODO add godoc
func Info(args ...interface{}) {
	logger.Info(args...)
}

//TODO add godoc
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

//TODO add godoc
func Error(args ...interface{}) {
	logger.Error(args...)
}

//TODO add godoc
func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

//TODO add godoc
func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

//TODO add godoc
func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

//TODO add godoc
func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

//TODO add godoc
func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

//TODO add godoc
func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

//TODO add godoc
func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

//TODO add godoc
func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}
